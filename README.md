# coding
使用golang设计一个简单的TCP服务端/客户端程序.conf/Server.xml有相应配置文件的说明。
1、服务器端通过stdin接收外部命令,服务器端接受"open a b c"这样的命令输入，将名为a,b,c的三个客户端的状态置为"open"，（客户端状态默认为"close"），相应的，也有"close"命令。open和close命令均需客户端回复后才生效,服务器端接受"status"命令，列出所有注册过的客户端，它们是否在线，它们的ip，它们的状态。
servermain.go 为服务器主函数入口，clientmain.go为客户端主函数入口。
2、go build -o serverout  编译得到服务器可执行程序。
3、go build -tags="client" -o clientout clientmain.go  编译得到客户端可执行程序。
4、./serverout 启动服务器程序。
5、./clientout ip name, ip 为服务器地址,格式如：127.0.0.1:8902，name为注册的客户端名称。
6、在服务器stdin输入status可以显示注册的客户端信息。
7、在服务器stdin输入open/close name1 name2 name3 ...可以打开或关闭注册客户端的状态。
