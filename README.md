## GitAudit

gitlab代码拉取审计


```shell
[root@tc git-audit]# ./git-audit -h
Usage of ./git-audit:
使用说明
  -v, --version          Display build and version msg.
  -c, --cfgFile string   Path to config yaml file.


Some errors have occurred, check and try again !!! 
```

默认配置文件： ~/.git-audit.yaml
```shell
database:
  type: mysql
  addr: xxxxx
  port: 3306
  dbname: xxxxx
  username: root
  password: xxxxx
gitLog:
  logLevel: debug
gitAudit:
  # feishu
  webhook: xxxxxxxx
  secret: xxxxxxx
  at: "<at user_id=\"ou_d1b9e980dad90457xxxxxxxxxx25\">名字</at>"
  sshFile: "/root/github/git-audit/test.log"
  httpFile: "/root/github/git-audit/http.log"
``
