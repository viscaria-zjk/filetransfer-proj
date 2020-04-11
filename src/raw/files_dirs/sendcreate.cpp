#define _CRT_SECURE_NO_WARNINGS
#include<stdio.h>
#include <io.h>
#include <direct.h>
#include <windows.h>
#include<stdlib.h>
#include <sys/stat.h>
#define IS_FOLDER 1
#define IS_FILE 2
#define NOT_EXIST 3
#define CREATE_SUCCESS 1
#define CREATE_FAIL 2
#define input_error 3

int judgePath(const char path[]) {
    struct _stat buf;
    /*  result:为-1时代表不存在，为0时代表存在  */
    int result = _stat(path, &buf);
    int type = 0;
    if (_S_IFDIR & buf.st_mode) {
        type = IS_FOLDER;
    }
    else if (_S_IFREG & buf.st_mode) {
        type = IS_FILE;
    }
    else
        type = NOT_EXIST;
    return type;
}
int sendcreate( char path[])
{   char PATH[200];
    
strcpy(PATH,path);
static int len=strlen(PATH);
	long handle;  //用于查找的句柄
	int i;
	struct _finddata_t fileinfo;
  static int DIR_Num=0,FIL_Num=0;
    FILE *fp;
  fp=fopen("send.txt","a");
  char * to_search=PATH;
  strcat(to_search,"/*");
  if (fp==0) 
	{
		printf("can't open file\n");
		return 0;
	}
  handle=_findfirst(to_search,&fileinfo);                          //第一次查找
  if(-1==handle)
		return -1; 
  _findnext(handle,&fileinfo);
	while(!_findnext(handle,&fileinfo))                              //循环查找其他符合的文件，直到找不到其他的为止
	{  char PATH[200];
		strcpy(PATH,path);
  strcat(PATH,"/");
  strcat(PATH,fileinfo.name);
 
	 if(judgePath(PATH)==IS_FILE)
   {FIL_Num++;
	 fp=fopen("send.txt","a");
   fprintf(fp,"FIL<%d><%s>\r",FIL_Num,fileinfo.name);
   fclose(fp);
   }

	}
	 handle=_findfirst(to_search,&fileinfo);                          //第一次查找
  if(-1==handle)
		return -1; 
  _findnext(handle,&fileinfo);
  while(!_findnext(handle,&fileinfo))                              //循环查找其他符合的文件，直到找不到其他的为止
	{  char PATH[200];
		strcpy(PATH,path);
  strcat(PATH,"/");
  strcat(PATH,fileinfo.name);
  if(judgePath(PATH)==IS_FOLDER)
		{DIR_Num++;
   fp=fopen("send.txt","a");
		fprintf(fp,"DIR<%d><%s>\r",DIR_Num,PATH+len+1);
		fclose(fp);
		sendcreate(PATH);
		}
	}
	_findclose(handle);                                              //关闭句柄
	
	
	
	return 0;

}
int main(){
	char path[200]={0};
	printf("请输入宁要整理的文件或文件夹");
	gets_s(path,200);
	sendcreate(path);
}

