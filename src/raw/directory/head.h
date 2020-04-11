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



/*
    函数judgePath():
    根据用户输入的本机路径，判断该路径是目录（directory）还是文件（file），或者根本不存在该目录或文件：
    返回值可以是：1 目录；2 文件；3 该地址不存在
    int judgePath(const char path[]);
*/


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

/*
   函数makeDir():
   按照Windows下的相对目录，在提供的根目录下递归创建对应文件夹
   返回1：创建成功；
   返回2：因权限不足或其他原因而创建失败
   int makeDir(const char relative[], const char root[]);

*/
int makeDir(const char relative[], const char root[]) {
    char path[200];
    char dir[200];//支持最大200字节绝对路径的输入
    strcpy(dir, root);
    FILE* fp;
    sprintf(path, "mkdir %s\\%s", root, relative);
    system(path);//创建目录
    strcat(dir, "\\");
    strcat(dir, relative);
    int result = judgePath(dir);//判断目录是否存在
    int signal = CREATE_FAIL;
    if (result == IS_FOLDER) {
        signal = CREATE_SUCCESS;
    }

    return signal;
}
