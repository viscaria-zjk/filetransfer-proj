
#include"head.h"




/*  judgePath()函数、makeDir函数使用测试  */
void main()
{
    printf("judgePath()函数使用测试：\n");
    printf("请输入文件/文件夹路径：（格式：D:/Program Files/OpenCV/imgs.jpg 或 D:\\\\Program Files\\\\OpenCV\\\\imgs.jpg 或 D:\\Program Files\\OpenCV\\imgs.jpg）\n");
    //printf("当前支持最大路径长度：200字节\n");
    char path[200] = { 0 };
    gets_s(path, 200);
    int result = judgePath(path);

    switch (result)
    {
    case 1:printf("It's a folder!\n"); break;
    case 2:printf("It's a file!\n"); break;
    case 3:printf("This path is not exist!\n"); break;
    default:
        break;
    }
    printf("makeDir()函数使用测试：\n");
    char relative[80] = { 0 }, root[80] = { 0 };
    printf("请输入相对路径：(格式： hwmm\\\\foo，当前支持最大路径长度：80字节）\n");
    gets_s(relative, 80);
    printf("请输入绝对路径：(格式： D:\\\\recv， 当前支持最大路径长度：80字节）\n");
    gets_s(root, 80);
    result = makeDir(relative, root);
    if (result == CREATE_SUCCESS)
        printf("创建成功！\n");
    else
        printf("创建失败!\n");
   


}