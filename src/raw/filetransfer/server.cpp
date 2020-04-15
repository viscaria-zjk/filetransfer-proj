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
DWORD WINAPI sendName(LPVOID p)
#else
void* sendName(void* arg)
#endif
{
	int iAddrSize = sizeof(remoteAddr);
	acRecvStr[0] = '\0';
	// ��������
	printf("%s\n", "���ڽ���������Ϣ��");
	recvfrom(iUDP, acRecvStr, 5, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
	// ÿ�ν���������Ϣ��ȴ�һ��ʱ��
	sleepUDP(10);
	// ���������Ϣ��ȷ�����ļ���
	if (acRecvStr[0] == 'n' && acRecvStr[1] == 'a' && acRecvStr[2] == 'm' && acRecvStr[3] == 'e' && acRecvStr[4] == '\0')
		sendto(iUDP, acFileName, iNameSize, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
	return 0;
}

// �����ļ�
void sendFile(char* acDirAndFileName, char* acIpAddr)
{
	int i, j, iFileSize, iSendNum, iAddrSize = sizeof(remoteAddr);
	FILE* pFile = NULL;

	pFile = fopen(acDirAndFileName, "rb");
	fseek(pFile, 0, SEEK_END);
	// �ļ��ֽ���
	iFileSize = ftell(pFile);
	intToStr(iFileSize, acFileSize);
	//printf("%s\n", acDirAndFileName);

	// ��ȡ�ļ�������
	iSize = strlen(acDirAndFileName);
	for (i = iSize - 1, iNameSize = 0; i >= 0; i--, iNameSize++)
		if (acDirAndFileName[i] == '\\' || acDirAndFileName[i] == '/')
			break;
	//printf("%d\n", iNameSize);
	// ��ȡ�ļ���
	for (i = 0; i < iNameSize; i++)
		acFileName[i] = acDirAndFileName[iSize - iNameSize + i];
	acFileName[iNameSize] = '\0';
	//printf("%s\n", acFileName);
	openUDP(acIpAddr);
	// �����ļ���
	for (;;)
	{
		// �����߳�
#ifdef _MSC_VER
		HANDLE hThread;
		DWORD threadId;
		hThread = CreateThread(NULL, 0, sendName, 0, 0, &threadId);
		// ÿ�ν���������Ϣ��ȴ�һ��ʱ��
		sleepUDP(1000);
		// ǿ����ֹ�߳�
		TerminateThread(hThread, 0);
#else
		pthread_t thread;
		void* thread_arg = (pthread_t)0;
		pthread_create(&thread, NULL, sendName, (void*)& thread_arg);
		// ÿ�ν���������Ϣ��ȴ�һ��ʱ��
		sleepUDP(1000);
		// ǿ����ֹ�߳�
		pthread_cancel(thread);
#endif
		// ���������Ϣ��ȷ�˳�ѭ��
		if (acRecvStr[0] == 'n' && acRecvStr[1] == 'a' && acRecvStr[2] == 'm' && acRecvStr[3] == 'e' && acRecvStr[4] == '\0')
			break;
	}
	// �����ļ��ֽ���
	for (;;)
	{
		acRecvStr[0] = '\0';
		// ��������
		recvfrom(iUDP, acRecvStr, 5, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
		// ÿ�ν���������Ϣ��ȴ�һ��ʱ��
		sleepUDP(10);
		// ���������Ϣ��ȷ
		if (acRecvStr[0] == 's' && acRecvStr[1] == 'i' && acRecvStr[2] == 'z' && acRecvStr[3] == 'e' && acRecvStr[4] == '\0')
		{
			// �����ļ��ֽ���
			sendto(iUDP, acFileSize, strlen(acFileSize), 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			break;
		}
	}

	iSendNum = iFileSize / SIZEB;
	// �����ļ�
	if (iSendNum > 0)
	{
		for (i = 0;; i++)
		{
			acRecvStr[0] = '\0';
			// ��������
			recvfrom(iUDP, acRecvStr, SIZEB, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
			printf("%s\t���ڷ����ļ��ĵ� %d ��\n", acRecvStr, i);
			// ÿ�ν���������Ϣ��ȴ�һ��ʱ��
			sleepUDP(10);
			fseek(pFile, strToInt(acRecvStr) * SIZEB, SEEK_SET);
			fread(acSendStr, 1, SIZEB, pFile);
			//printf("%s\n", acSendStr);
			// ����һ���ļ�
			sendto(iUDP, acSendStr, SIZEB, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
			if (strToInt(acRecvStr) >= iSendNum - 1)
				break;
		}
	}
	// �����ļ�ʣ���ֽ�
	iSize = iFileSize % SIZEB;
	if (iSize > 0)
	{
		for (;;)
		{
			acRecvStr[0] = '\0';
			// ��������
			recvfrom(iUDP, acRecvStr, 5, 0, (struct sockaddr*) & remoteAddr, &iAddrSize);
			// ÿ�ν���������Ϣ��ȴ�һ��ʱ��
			sleepUDP(10);
			// ���������Ϣ��ȷ
			if (acRecvStr[0] == 'l' && acRecvStr[1] == 'a' && acRecvStr[2] == 's' && acRecvStr[3] == 't' && acRecvStr[4] == '\0')
			{
				fseek(pFile, iSendNum * SIZEB, SEEK_SET);
				fread(acSendStr, 1, iSize, pFile);
				//printf("%s\n", acSendStr);
				// �����ļ�ʣ���ֽ�
				sendto(iUDP, acSendStr, iSize, 0, (struct sockaddr*) & clientAddr, sizeof(clientAddr));
				break;
			}
		}
	}
	printf("%s\n", "�ļ�������ϣ�");

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
	printf("%s\n", "������Ϊ�������ˣ����ڷ����ļ���");
	iServerPort = 1025;
	iClientPort = 1024;
	fileName:
		printf("%s\n", "��������Ҫ���͵�·���ļ���:");
		scanf("%s", acDirAndFileName);
		pFile = fopen(acDirAndFileName, "rb");
		if (pFile == NULL)
		{
			printf("%s\n", "��ȡ�ļ�ʧ�ܣ������������ļ�����");
			goto fileName;
		}
		// �ر��ļ�
		fclose(pFile);
		printf("%s\n", "����������ļ����� IP ��ַ�����磺192.168.101.6");
		scanf("%s", acIpAddr);
		sendFile(acDirAndFileName, acIpAddr);

	return 0;
}