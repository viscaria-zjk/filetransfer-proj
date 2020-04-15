#define  _CRT_SECURE_NO_WARNINGS
#pragma warning(disable:4996)
#include<stdio.h>
#include<stdlib.h>
#include<string.h>
#include<math.h>

#ifdef _MSC_VER
#include<winsock2.h>
#include<windows.h>
#pragma comment(lib, "ws2_32.lib")
#else
#include<pthread.h>
#include<unistd.h>
#include<signal.h>
#include<sys/socket.h>
#include<arpa/inet.h>

#endif

// 存放发送接收字符数组大小
#define SIZEA 65501
// 每次发送接收字节数
#define SIZEB 65500

typedef struct sockaddr_in SockAddrIn;
SockAddrIn serverAddr, remoteAddr, clientAddr;

// 端口号
int iServerPort, iClientPort;
// 新建 socket 信息
int iUDP;

// 字符串转整型
int strToInt(char* acStr)
{
	int i, iIndex = 0, iNum = 0, iSize = 0;
	if (acStr[0] == '+' || acStr[0] == '-')
		iIndex = 1;

	for (iSize = iIndex; ; iSize++)
		if (acStr[iSize] < '0' || acStr[iSize] > '9')
			break;

	for (i = iIndex; i < iSize; i++)
		iNum += (int)pow(10, iSize - i - 1) * (acStr[i] - 48);

	if (acStr[0] == '-')
		iNum = -iNum;

	return iNum;
}

// 整型转字符串
void intToStr(int iInt, char* acStr)
{
	int iIndex = 0, iSize, iNum, iBit, i, j;

	if (iInt < 0)
	{
		acStr[0] = '-';
		iInt = -iInt;
		iIndex = 1;
	}
	for (i = 0; ; i++)
		if (iInt < pow(10, i))
			break;
	iSize = i;

	for (i = 0; i < iSize; i++)
	{
		iNum = pow(10, iSize - i - 1);
		iBit = iInt / iNum;
		iInt -= iNum * iBit;
		acStr[i + iIndex] = iBit + 48;
	}
	if (iSize != 0)
		acStr[iSize + iIndex] = '\0';
	else
	{
		acStr[0] = '0';
		acStr[1] = '\0';
	}
}

void sleepUDP(int iSleep)
{
#ifdef _MSC_VER
	Sleep(iSleep);
#else
	usleep(iSleep * 1000);
#endif
}

void openUDP(char* acIpAddr)
{
#ifdef _MSC_VER
	// Winsows 启用 socket
	WSADATA wsadata;
	if (WSAStartup(MAKEWORD(1, 1), &wsadata) == SOCKET_ERROR)
	{
		printf("启用 socket 失败\n");
		exit(0);
	}
#endif

	// 新建 socket
	if ((iUDP = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP)) == -1)
	{
		printf("新建 socket 失败\n");
		exit(0);
	}

	// 清零
	memset(&serverAddr, 0, sizeof(serverAddr));
	memset(&clientAddr, 0, sizeof(clientAddr));

	// 设置协议 IP 地址及 Port
	serverAddr.sin_family = AF_INET;
	serverAddr.sin_port = htons(iServerPort);
	serverAddr.sin_addr.s_addr = htonl(INADDR_ANY);

	clientAddr.sin_family = AF_INET;
	clientAddr.sin_port = htons(iClientPort);
	clientAddr.sin_addr.s_addr = inet_addr(acIpAddr);

	// 绑定端口，监听端口
	if (bind(iUDP, (struct sockaddr*) & serverAddr, sizeof(serverAddr)) == -1)
	{
		printf("绑定端口失败\n");
		exit(0);
	}
}

void closeUDP(void)
{
#ifdef _MSC_VER
	// Winsows 关闭 socket
	closesocket(iUDP);
	WSACleanup();
#endif
}

// 要发送的字符串
char acSendStr[SIZEA];
// 接收到的字符串
char acRecvStr[SIZEA];
// 请求信息
char acReq[SIZEA];
// 文件名字符串
char acFileName[SIZEA];
// 文件字节数字符串
char acFileSize[SIZEA];

int iSize, iNameSize;


// 发送文件名
#ifdef _MSC_VER
DWORD WINAPI sendName(LPVOID p)
#else
void* sendName(void* arg)
#endif
{
	int iAddrSize = sizeof(remoteAddr);
	acRecvStr[0] = '\0';
	// 接收请求
	printf("%s\n", "正在接收请求信息！");
	recvfrom(iUDP, acRecvStr, 5, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
	// 每次接收请求信息后等待一段时间
	sleepUDP(10);
	// 如果请求信息正确发送文件名
	if (acRecvStr[0] == 'n' && acRecvStr[1] == 'a' && acRecvStr[2] == 'm' && acRecvStr[3] == 'e' && acRecvStr[4] == '\0')
		sendto(iUDP, acFileName, iNameSize, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
	return 0;
}

// 发送文件
void sendFile(char* acDirAndFileName, char* acIpAddr)
{
	int i, j, iFileSize, iSendNum, iAddrSize = sizeof(remoteAddr);
	FILE* pFile = NULL;

	pFile = fopen(acDirAndFileName, "rb");
	fseek(pFile, 0, SEEK_END);
	// 文件字节数
	iFileSize = ftell(pFile);
	intToStr(iFileSize, acFileSize);
	//printf("%s\n", acDirAndFileName);

	// 获取文件名长度
	iSize = strlen(acDirAndFileName);
	for (i = iSize - 1, iNameSize = 0; i >= 0; i--, iNameSize++)
		if (acDirAndFileName[i] == '\\' || acDirAndFileName[i] == '/')
			break;
	//printf("%d\n", iNameSize);
	// 截取文件名
	for (i = 0; i < iNameSize; i++)
		acFileName[i] = acDirAndFileName[iSize - iNameSize + i];
	acFileName[iNameSize] = '\0';
	//printf("%s\n", acFileName);
	openUDP(acIpAddr);
	// 发送文件名
	for (;;)
	{
		// 创建线程
#ifdef _MSC_VER
		HANDLE hThread;
		DWORD threadId;
		hThread = CreateThread(NULL, 0, sendName, 0, 0, &threadId);
		// 每次接收请求信息后等待一段时间
		sleepUDP(1000);
		// 强制终止线程
		TerminateThread(hThread, 0);
#else
		pthread_t thread;
		void* thread_arg = (pthread_t)0;
		pthread_create(&thread, NULL, sendName, (void*)& thread_arg);
		// 每次接收请求信息后等待一段时间
		sleepUDP(1000);
		// 强制终止线程
		pthread_cancel(thread);
#endif
		// 如果请求信息正确退出循环
		if (acRecvStr[0] == 'n' && acRecvStr[1] == 'a' && acRecvStr[2] == 'm' && acRecvStr[3] == 'e' && acRecvStr[4] == '\0')
			break;
	}
	// 发送文件字节数
	for (;;)
	{
		acRecvStr[0] = '\0';
		// 接收请求
		recvfrom(iUDP, acRecvStr, 5, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
		// 每次接收请求信息后等待一段时间
		sleepUDP(10);
		// 如果请求信息正确
		if (acRecvStr[0] == 's' && acRecvStr[1] == 'i' && acRecvStr[2] == 'z' && acRecvStr[3] == 'e' && acRecvStr[4] == '\0')
		{
			// 发送文件字节数
			sendto(iUDP, acFileSize, strlen(acFileSize), 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			break;
		}
	}

	iSendNum = iFileSize / SIZEB;
	// 发送文件
	if (iSendNum > 0)
	{
		for (i = 0;; i++)
		{
			acRecvStr[0] = '\0';
			// 接收请求
			recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
			printf("%s\t正在发送文件的第 %d 段\n", acRecvStr, i);
			// 每次接收请求信息后等待一段时间
			sleepUDP(10);
			fseek(pFile, strToInt(acRecvStr) * SIZEB, SEEK_SET);
			fread(acSendStr, 1, SIZEB, pFile);
			//printf("%s\n", acSendStr);
			// 发送一段文件
			sendto(iUDP, acSendStr, SIZEB, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			if (strToInt(acRecvStr) >= iSendNum - 1)
				break;
		}
	}
	// 发送文件剩余字节
	iSize = iFileSize % SIZEB;
	if (iSize > 0)
	{
		for (;;)
		{
			acRecvStr[0] = '\0';
			// 接收请求
			recvfrom(iUDP, acRecvStr, 5, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
			// 每次接收请求信息后等待一段时间
			sleepUDP(10);
			// 如果请求信息正确
			if (acRecvStr[0] == 'l' && acRecvStr[1] == 'a' && acRecvStr[2] == 's' && acRecvStr[3] == 't' && acRecvStr[4] == '\0')
			{
				fseek(pFile, iSendNum * SIZEB, SEEK_SET);
				fread(acSendStr, 1, iSize, pFile);
				//printf("%s\n", acSendStr);
				// 发送文件剩余字节
				sendto(iUDP, acSendStr, iSize, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
				break;
			}
		}
	}
	printf("%s\n", "文件发送完毕！");

	// 关闭连接
	closeUDP();
}

int main(void)
{
	char acDirName[SIZEA];
	char acDirAndFileName[SIZEA];
	char acIpAddr[15];
	int i, iOption = 0, iSize = 0;
	FILE* pFile = NULL;
	char cLast = '\\';
	printf("%s\n", "本程序为服务器端，用于发送文件。");
	iServerPort = 1025;
	iClientPort = 1024;
	fileName:
		printf("%s\n", "请输入需要发送的路径文件名:");
		scanf("%s", acDirAndFileName);
		pFile = fopen(acDirAndFileName, "rb");
		if (pFile == NULL)
		{
			printf("%s\n", "读取文件失败，请重新输入文件名。");
			goto fileName;
		}
		// 关闭文件
		fclose(pFile);
		printf("%s\n", "请输入接收文件方的 IP 地址，例如：192.168.101.6");
		scanf("%s", acIpAddr);
		sendFile(acDirAndFileName, acIpAddr);

	return 0;
}