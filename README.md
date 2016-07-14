Phoenix-打点服务
============

以mongodb 固定集合作为持久存储. golang实现的打点服务。 目前主要用在线上API性能监控等。

运行即可


    $ ./phoenix -host 1.1.1.1 -port 8888 -cs 127.0.0.1:27017


###支持的参数：

Usage of ./phoenix:

  -cs string

    	Mongodb Connection String. (default "192.168.1.102:27017")

  -host string

    	bound ip. default:localhost (default "localhost")

  -port string

    	port. default:8888 (default "8888")

