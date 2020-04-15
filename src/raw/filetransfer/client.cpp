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

// ��ŷ��ͽ����ַ������С
#define SIZEA 65501
// ÿ�η��ͽ����ֽ���
#define SIZEB 65500

typedef struct sockaddr_in SockAddrIn;
SockAddrIn serverAddr, remoteAddr, clientAddr;

// �˿ں�
int iServerPort, iClientPort;
// �½� socket ��Ϣ
int iUDP;

// �ַ���ת����
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

// ����ת�ַ���
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
	// Winsows ���� socket
	WSADATA wsadata;
	if (WSAStartup(MAKEWORD(1, 1), &wsadata) == SOCKET_ERROR)
	{
		printf("���� socket ʧ��\n");
		exit(0);
	}
#endif

	// �½� socket
	if ((iUDP = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP)) == -1)
	{
		printf("�½� socket ʧ��\n");
		exit(0);
	}

	// ����
	memset(&serverAddr, 0, sizeof(serverAddr));
	memset(&clientAddr, 0, sizeof(clientAddr));

	// ����Э�� IP ��ַ�� Port
	serverAddr.sin_family = AF_INET;
	serverAddr.sin_port = htons(iServerPort);
	serverAddr.sin_addr.s_addr = htonl(INADDR_ANY);

	clientAddr.sin_family = AF_INET;
	clientAddr.sin_port = htons(iClientPort);
	clientAddr.sin_addr.s_addr = inet_addr(acIpAddr);

	// �󶨶˿ڣ������˿�
	if (bind(iUDP, (struct sockaddr*) & serverAddr, sizeof(serverAddr)) == -1)
	{
		printf("�󶨶˿�ʧ��\n");
		exit(0);
	}
}

void closeUDP(void)
{
#ifdef _MSC_VER
	// Winsows �ر� socket
	closesocket(iUDP);
	WSACleanup();
#endif
}

// Ҫ���͵��ַ���
char acSendStr[SIZEA];
// ���յ����ַ���
char acRecvStr[SIZEA];
// ������Ϣ
char acReq[SIZEA];
// �ļ����ַ���
char acFileName[SIZEA];
// �ļ��ֽ����ַ���
char acFileSize[SIZEA];

int iSize, iNameSize;

// �����ļ���
#ifdef _MSC_VER
DWORD WINAPI recvName(LPVOID p)
#else
void* recvName(void* arg)
#endif
{
	int iAddrSize = sizeof(remoteAddr);
	acReq[0] = 'n'; acReq[1] = 'a'; acReq[2] = 'm'; acReq[3] = 'e'; acReq[4] = '\0';
	acRecvStr[0] = '\0';
	printf("%s\n", "���ڷ���������Ϣ��");
	// ����������Ϣ
	sendto(iUDP, acReq, strlen(acReq), 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
	// ÿ�η���������Ϣ��ȴ�һ��ʱ��
	sleepUDP(10);
	// �����ļ���
	iSize = recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
	return 0;
}

// �����ļ�
void recvFile(char* acDirName, char* acIpAddr)
{
	FILE* pFile = NULL;
	int i, iFileSize, iRecvNum, iAddrSize = sizeof(remoteAddr);
	// ·���ļ���
	char acDirAndFileName[SIZEA];

	openUDP(acIpAddr);
	// �����ļ���
	for (;;)
	{
		// �����߳�
#ifdef _MSC_VER
		HANDLE hThread;
		DWORD threadId;
		hThread = CreateThread(NULL, 0, recvName, 0, 0, &threadId);
		// ÿ�η��ͺ�ȴ�һ��ʱ��
		sleepUDP(1000);
		// ǿ����ֹ�߳�
		TerminateThread(hThread, 0);
#else
		pthread_t thread;
		void* thread_arg = (pthread_t)0;
		pthread_create(&thread, NULL, recvName, (void*)& thread_arg);
		// ÿ�η��ͺ�ȴ�һ��ʱ��
		sleepUDP(1000);
		// ǿ����ֹ�߳�
		pthread_cancel(thread);
#endif
		if (acRecvStr[0] != '\0')
		{
			acRecvStr[iSize] = '\0';
			printf("�ļ���Ϊ��%s\n", acRecvStr);
			break;
		}
	}

	acDirAndFileName[0] = '\0';
	strcat(acDirAndFileName, acDirName);
	// ����·�������ļ���
	strcat(acDirAndFileName, acRecvStr);
	// ����Ѿ�������ļ��˾�����ļ�����
	pFile = fopen(acDirAndFileName, "w");
	fclose(pFile);

	acReq[0] = 's'; acReq[1] = 'i'; acReq[2] = 'z'; acReq[3] = 'e'; acReq[4] = '\0';
	// �����ļ��ֽ���
	for (;;)
	{
		// ����������Ϣ
		sendto(iUDP, acReq, strlen(acReq) + 1, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
		// ÿ�η���������Ϣ��ȴ�һ��ʱ��
		sleepUDP(10);
		// �����ļ��ֽ���
		acRecvStr[0] = '\0';
		iSize = recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
		if (acRecvStr[0] != '\0')
		{
			acRecvStr[iSize] = '\0';
			iFileSize = strToInt(acRecvStr);
			printf("�ļ��ֽ���Ϊ��%d\n", iFileSize);
			break;
		}
	}

	// ��׷�ӷ�ʽд���ļ�
	pFile = fopen(acDirAndFileName, "ab");
	// �ļ��ּ��ν���
	iRecvNum = iFileSize / SIZEB;
	// �����ļ�
	for (i = 0; i < iRecvNum; i++)
	{
		intToStr(i, acReq);
		for (;;)
		{
			// ����������Ϣ
			sendto(iUDP, acReq, strlen(acReq) + 1, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			printf("%s\t���ڽ����ļ��ĵ� %d ��\n", acReq, i);
			// ÿ�η���������Ϣ��ȴ�һ��ʱ��
			sleepUDP(10);
			// ����һ���ļ�
			iSize = recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
			if (iSize == SIZEB)
			{
				// ��׷�ӷ�ʽд���ļ�
				fwrite(acRecvStr, sizeof(char), iSize, pFile);
				break;
			}
		}
	}
	// �����ļ�ʣ���ֽ�
	iSize = iFileSize % SIZEB;
	if (iSize > 0)
	{
		acReq[0] = 'l'; acReq[1] = 'a'; acReq[2] = 's'; acReq[3] = 't'; acReq[4] = '\0';
		for (;;)
		{
			// ����������Ϣ
			sendto(iUDP, acReq, strlen(acReq) + 1, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			// ÿ�η���������Ϣ��ȴ�һ��ʱ��
			sleepUDP(10);
			// �����ļ�ʣ���ֽ�
			if (recvfrom(iUDP, acRecvStr, iSize, 0, (struct sockaddr*) & remoteAddr, &iAddrSize) == iSize)
			{
				// ��׷�ӷ�ʽд���ļ�
				fwrite(acRecvStr, sizeof(char), iSize, pFile);
				break;
			}
		}
	}
	printf("%s\n", "�ļ�������ϣ�");

	// �ر��ļ�
	fclose(pFile);
	// �ر�����
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
	printf("%s\n", "������Ϊ�ͻ��ˣ����ڽ����ļ���");
	iServerPort = 1024;
	iClientPort = 1025;
	dirName:
		printf("%s\n", "�����뱣���ļ���·����:");
		scanf("%s", acDirName);
		iSize = strlen(acDirName);
		// ����ǲ��� Linux ·����
		for (i = 0; i < iSize; i++)
		{
			if (acDirName[i] == '/')
			{
				cLast = '/';
				break;
			}
		}
		// ���·�������һ���ַ��ǲ��� \ �� /
		if (acDirName[iSize - 1] != cLast)
		{
			acDirName[iSize] = cLast;
			acDirName[iSize + 1] = '\0';
		}
		acDirAndFileName[0] = '\0';
		strcat(acDirAndFileName, acDirName);
		strcat(acDirAndFileName, "a.txt");
		// ��̽����һ���޹ؽ�Ҫ���ļ�
		pFile = fopen(acDirAndFileName, "w");
		if (pFile == NULL)
		{
			printf("%s\n", "��·���޷������ļ�������������·������");
			goto dirName;
		}
		else
		{
			// �ر��ļ�
			fclose(pFile);
			// ɾ���ļ�
			remove(acDirAndFileName);
		}

		printf("%s\n", "�����뷢���ļ����� IP ��ַ�������пո����磺192.168.101.6");
		scanf("%s", acIpAddr);
		recvFile(acDirName, acIpAddr);

	return 0;
}