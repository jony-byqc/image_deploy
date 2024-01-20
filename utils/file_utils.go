package utils

import (
	"archive/tar"
	"compress/gzip"
	log "github.com/sirupsen/logrus"
	"github.com/vbauerster/mpb/v7"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func IsFile(path string) bool {

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("文件不存在")
		} else {
			log.Errorf("无法获取文件信息：%v\n", err)
		}
		return false
	}

	return fileInfo.Mode().IsRegular()
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CopyFile(dstName, srcName string, bar *mpb.Bar) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return
	}
	defer dst.Close()

	if bar != nil {
		return io.Copy(dst, bar.ProxyReader(src))
	}
	return io.Copy(dst, src)
}

func UnGzip(src, dest string) error {
	srcFn, err := os.OpenFile(src, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	defer srcFn.Close()

	destFn, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	defer destFn.Close()

	r, err := gzip.NewReader(srcFn)
	if err != nil {
		return err
	} else {
		defer r.Close()
		_, err = io.Copy(destFn, r)
		if err != nil {
			return err
		}
		return nil
	}

}

func TarGzWrite(_dpath, _spath string, tw *tar.Writer, fi os.FileInfo) error {
	fr, err := os.Open(_dpath + "/" + _spath)
	if err != nil {
		return err
	}
	defer fr.Close()

	h := new(tar.Header)

	h.Name = _spath
	h.Size = fi.Size()
	h.Mode = int64(fi.Mode())
	h.ModTime = fi.ModTime()
	err = tw.WriteHeader(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, fr)
	if err != nil {
		return err
	}
	return nil
}

func IterDirectory(dirPath, subpath string, tw *tar.Writer) error {
	dir, err := os.Open(dirPath + "/" + subpath)
	if err != nil {
		return err
	}
	defer dir.Close()
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		var curpath string
		if subpath == "" {
			curpath = fi.Name()
		} else {
			curpath = subpath + "/" + fi.Name()
		}

		if fi.IsDir() {
			err := IterDirectory(dirPath, curpath, tw)
			if err != nil {
				return err
			}
		} else {
			err := TarGzWrite(dirPath, curpath, tw, fi)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func TarGz(outFilePath string, inPath string) error {
	inPath = strings.TrimRight(inPath, "/")
	// file write
	fw, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer fw.Close()

	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = IterDirectory(inPath, "", tw)
	if err != nil {
		return err
	}

	return nil
}

type DockerCompose struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Image string `yaml:"image"`
}

func FindCompose() (images, paths []string) {
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("遍历目录时发生错误:", err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.Contains(info.Name(), "docker-compose.y") { // yaml/yml
			images = append(images, FindImagesInCompose(path)...)
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		log.Fatal("遍历目录时发生错误:", err)
	}
	return
}

func FindImagesInCompose(path string) []string {
	images := make([]string, 0)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("无法读取文件 %s: %v", path, err)
	}

	var dockerCompose DockerCompose
	err = yaml.Unmarshal(content, &dockerCompose)
	if err != nil {
		log.Fatalf("无法解析 YAML 文件: %v", err)
	}

	for _, service := range dockerCompose.Services {
		images = append(images, service.Image)
	}
	return images
}
