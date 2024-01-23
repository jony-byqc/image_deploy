# image_deploy
docker image automatically deploy to machine

deployment.yaml
```yaml
user:   # docker 仓库用户名
password: # docker 仓库token
type: PullAndDeploy # 三种部署方式pull；deploy；pullAndDeploy
ssh_config:
  - ip: 192.168.60.157
    port: 22
    user: wjq
    password: 123456
deploy:
  target_folder: /compose # 将同级目录下所有文件上传至此目录
```

1.将同级目录下所有文件按照相同目录级别上传到所需部署的机器中
2.解析同级目录下所有docker-compose文件 images
3.执行命令自动运行docker-compose