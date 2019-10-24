### 微服务进程管理工具

#### 用途
微服务很多的时候要一个个手动启动比较麻烦，该工具作为父进程，把某一个目录下的程序作为  
子进程启动起来，并对目录进行监控。当增加，重命名或删除目录下的程序文件时，会自动启动，  
重启或停止相关子进程

#### 使用
```shell
# /tmp/test目录下放程序文件
./admin_launcher dev /tmp/test localhost:2379
```

#### 注意
  - 当管理进程收到`SIGTERM`或`SIGQUIT`或`SIGINT`信号，比如kill $pid，管理进程和子进程都退出
  
  - 当管理进程收到`SIGKILL`信号，比如kill -9 $pid，管理进程被杀死，子进程不退出，父进程变为1号进程
    
  - 当管理进程收到`SIGTSTP`信号，比如ctrl + z，管理进程被挂起(`jobs`命令查看)，子进程暂停服务，执行`bg %N`或者`fg %N`恢复

#### 提示
  - 当父进程收到`SIGTERM`或SIGQUIT信号，比如kill $pid，父进程退出，子进程不退出，父进程变为1号进程

  - 当父进程收到`SIGINT`信号，比如ctrl + c，子进程也会收到信号，和父进程一起退出

  - 当父进程收到`SIGKILL`信号，比如kill -9 $pid，信号不能被捕获和忽略，父进程被杀死，子进程不退出，父进程变为1号进程
  
  - 当父进程收到`SIGTSTP`信号，比如ctrl + z，父进程被挂起(`jobs`命令查看)，子进程也是，执行`bg %N`或者`fg %N`恢复