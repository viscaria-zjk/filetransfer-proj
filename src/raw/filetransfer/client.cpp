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

// 接收文件名
#ifdef _MSC_VER
DWORD WINAPI recvName(LPVOID p)
#else
void* recvName(void* arg)
#endif
{
	int iAddrSize = sizeof(remoteAddr);
	acReq[0] = 'n'; acReq[1] = 'a'; acReq[2] = 'm'; acReq[3] = 'e'; acReq[4] = '\0';
	acRecvStr[0] = '\0';
	printf("%s\n", "正在发送请求信息！");
	// 发送请求信息
	sendto(iUDP, acReq, strlen(acReq), 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
	// 每次发送请求信息后等待一段时间
	sleepUDP(10);
	// 接收文件名
	iSize = recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
	return 0;
}

// 接收文件
void recvFile(char* acDirName, char* acIpAddr)
{
	FILE* pFile = NULL;
	int i, iFileSize, iRecvNum, iAddrSize = sizeof(remoteAddr);
	// 路径文件名
	char acDirAndFileName[SIZEA];

	openUDP(acIpAddr);
	// 接收文件名
	for (;;)
	{
		// 创建线程
#ifdef _MSC_VER
		HANDLE hThread;
		DWORD threadId;
		hThread = CreateThread(NULL, 0, recvName, 0, 0, &threadId);
		// 每次发送后等待一段时间
		sleepUDP(1000);
		// 强制终止线程
		TerminateThread(hThread, 0);
#else
		pthread_t thread;
		void* thread_arg = (pthread_t)0;
		pthread_create(&thread, NULL, recvName, (void*)& thread_arg);
		// 每次发送后等待一段时间
		sleepUDP(1000);
		// 强制终止线程
		pthread_cancel(thread);
#endif
		if (acRecvStr[0] != '\0')
		{
			acRecvStr[iSize] = '\0';
			printf("文件名为：%s\n", acRecvStr);
			break;
		}
	}

	acDirAndFileName[0] = '\0';
	strcat(acDirAndFileName, acDirName);
	// 连接路径名和文件名
	strcat(acDirAndFileName, acRecvStr);
	// 如果已经有这个文件了就清空文件内容
	pFile = fopen(acDirAndFileName, "w");
	fclose(pFile);

	acReq[0] = 's'; acReq[1] = 'i'; acReq[2] = 'z'; acReq[3] = 'e'; acReq[4] = '\0';
	// 接收文件字节数
	for (;;)
	{
		// 发送请求信息
		sendto(iUDP, acReq, strlen(acReq) + 1, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
		// 每次发送请求信息后等待一段时间
		sleepUDP(10);
		// 接收文件字节数
		acRecvStr[0] = '\0';
		iSize = recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
		if (acRecvStr[0] != '\0')
		{
			acRecvStr[iSize] = '\0';
			iFileSize = strToInt(acRecvStr);
			printf("文件字节数为：%d\n", iFileSize);
			break;
		}
	}

	// 以追加方式写入文件
	pFile = fopen(acDirAndFileName, "ab");
	// 文件分几次接收
	iRecvNum = iFileSize / SIZEB;
	// 接收文件
	for (i = 0; i < iRecvNum; i++)
	{
		intToStr(i, acReq);
		for (;;)
		{
			// 发送请求信息
			sendto(iUDP, acReq, strlen(acReq) + 1, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			printf("%s\t正在接收文件的第 %d 段\n", acReq, i);
			// 每次发送请求信息后等待一段时间
			sleepUDP(10);
			// 接收一段文件
			iSize = recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
			if (iSize == SIZEB)
			{
				// 以追加方式写入文件
				fwrite(acRecvStr, sizeof(char), iSize, pFile);
				break;
			}
		}
	}
	// 接收文件剩余字节
	iSize = iFileSize % SIZEB;
	if (iSize > 0)
	{
		acReq[0] = 'l'; acReq[1] = 'a'; acReq[2] = 's'; acReq[3] = 't'; acReq[4] = '\0';
		for (;;)
		{
			// 发送请求信息
			sendto(iUDP, acReq, strlen(acReq) + 1, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			// 每次发送请求信息后等待一段时间
			sleepUDP(10);
			// 接收文件剩余字节
			if (recvfrom(iUDP, acRecvStr, iSize, 0, (struct sockaddr*) & remoteAddr, &iAddrSize) == iSize)
			{
				// 以追加方式写入文件
				fwrite(acRecvStr, sizeof(char), iSize, pFile);
				break;
			}
		}
	}
	printf("%s\n", "文件接收完毕！");

	// 关闭文件
	fclose(pFile);
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
	printf("%s\n", "本程序为客户端，用于接收文件。");
	iServerPort = 1024;
	iClientPort = 1025;
	dirName:
		printf("%s\n", "请输入保存文件的路径名:");
		scanf("%s", acDirName);
		iSize = strlen(acDirName);
		// 检查是不是 Linux 路径名
		for (i = 0; i < iSize; i++)
		{
			if (acDirName[i] == '/')
			{
				cLast = '/';
				break;
			}
		}
		// 检查路径名最后一个字符是不是 \ 或 /
		if (acDirName[iSize - 1] != cLast)
		{
			acDirName[iSize] = cLast;
			acDirName[iSize + 1] = '\0';
		}
		acDirAndFileName[0] = '\0';
		strcat(acDirAndFileName, acDirName);
		strcat(acDirAndFileName, "a.txt");
		// 试探保存一个无关紧要的文件
		pFile = fopen(acDirAndFileName, "w");
		if (pFile == NULL)
		{
			printf("%s\n", "该路径无法创建文件，请重新输入路径名。");
			goto dirName;
		}
		else
		{
			// 关闭文件
			fclose(pFile);
			// 删除文件
			remove(acDirAndFileName);
		}

		printf("%s\n", "请输入发送文件方的 IP 地址，不能有空格。例如：192.168.101.6");
		scanf("%s", acIpAddr);
		recvFile(acDirName, acIpAddr);

	return 0;
}