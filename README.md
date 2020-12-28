# hearthstone-pf
炉石传说macOS拔线工具，需要系统版本10.9及以上

思路：  
1、找出炉石传说跟服务端的通讯端口  
2、通过Mac上自带的Packet Filter 防火墙block掉通讯端口，造成网络中断的假象，让炉石传说客户端发起重新连接  
3、去掉block规则，恢复通讯

参数说明:  
-d 			炉石传说断网  
-e 			炉石传说网络恢复  
-s 			自动重连间隔(单位秒)  
-debug    调试模式  
-b 			备份配置文件

使用示例(以文件放在下载目录为例):  
备份pf默认配置文件(会在程序目录下生成pf.conf_2020xxx的备份文件): sudo ~/Downloads/hearthstone-pf -b  
自动档(断网8秒后恢复网络):  sudo ~/Downloads/hearthstone-pf -e -s 8  
手动档(断网): 		   				sudo ~/Downloads/hearthstone-pf -e  
手动档(网络恢复): 	   			sudo ~/Downloads/hearthstone-pf -d  

