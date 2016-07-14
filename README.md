Phoenix-打点服务
============

以mongodb 固定集合作为持久存储. golang实现的打点服务。 目前主要用在线上API性能监控等。

###文件解释

##phoenix.php   
php的SDK

##phoenix_admin.go    
查看数据曲线图。  暂无权限控制，启动了 看完了 记得干掉进程。。。哈哈



运行即可


    $ ./phoenix -host 1.1.1.1 -port 8888 -cs 1.1.1.1:27017


###支持的参数：

Usage of ./phoenix:

  -cs string

    	Mongodb Connection String. (default "127.0.0.1:27017")

  -host string

    	bound ip. default:localhost (default "localhost")

  -port string

    	port. default:8888 (default "8888")

