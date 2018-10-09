package main
import (
	"os"
	"io"
	"fmt"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { // 文件或者目录存在
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//我们自定义要给函数，将一个文件拷贝到另外一个地方
//1. 先判断原来文件是否存在，如果原文件不存在直接返回
//2. 判断目标文件是否存在，如果不存在则创建，再拷贝
func MyCopy(dstFilePath string, srcFilePath string) (int64, error) {

	//1.先判断原来文件是否存在，如果原文件不存在直接返回
	if b, err := PathExists(srcFilePath); !b {
		fmt.Println("文件不存在或者其它错误 err=", err)
		return 0, err
	} 

	//2. 打开 srcFile
	srcFile , err := os.Open(srcFilePath)
	if err != nil {
		fmt.Println("open file err=", err)
		return 0, err
	}

	//3. 判断目标文件是否存在，如果不存在则创建，再拷=> 直接使用 
	dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY | os.O_CREATE , 0666 )
	if err != nil {
		fmt.Println("OpenFile  err=", err)
		return 0, err
	}
	//4. 调用Copy 完成拷贝
	return io.Copy(dstFile, srcFile)
}

func main() {

// 	说明：将一张图片/电影/mp3拷贝到另外一个文件  e:/abc.jpg   io包
// func io.Copy(dst Writer, src Reader) (written int64, err error)
// 1)如果dst文件不存在，则创建
// 2)自定义函数 CopyFile(dstFile, srcFile string) (written int64, err error)

	srcFile := "d:/尚硅谷_韩顺平_Go语言核心编程new.doc"
	dstFile := "f:/Go语言核心编程new.doc"
	_, err := MyCopy(dstFile, srcFile)
	if err != nil {
		fmt.Println("copy file 错误", err)
	}
	fmt.Println("copy success~~!!")
	

}
