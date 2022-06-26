# html-Absorber

使用标准输入，从URL或HTML文件中批量提取、筛选出指定标签内容，属性值及注释中的内容

**建议：Linux/Mac下使用效果最佳**

## Install

```
go install -v github.com/zha0gongz1/html-Absorber@latest
```

## Usage

```
▶ html-Absorber 
Usage: html-Absorber <mode> [<args>]
Modes:
 tags <tag names>        Extract text contained in tags
 attribs <attrib names>  Extract attribute values
 comments                Extract comments
Option:
 -output                 Save the result to file

Examples:
 cat urls.txt | html-Absorber tags title [-output]
 find . -type f -name "*.html" | html-Absorber attribs src href [-output]
 cat urls.txt | html-Absorber comments [-output]
```

## Demo
<center>
    <img style="width: 100%; border-radius: 0.32em;
    box-shadow: 0 2px 5px 0 rgba(35,36,38,.12),0 2px 10px 0 rgba(35,36,38,.08);" 
    src="https://raw.githubusercontent.com/zha0gongz1/html-Absorber/main/pic/1656234385060.jpg">
</center>

<p align="center">命令行输出结果</p>

<center>
    <img style="width: 100%; border-radius: 0.32em;
    box-shadow: 0 2px 5px 0 rgba(35,36,38,.12),0 2px 10px 0 rgba(35,36,38,.08);" 
    src="https://raw.githubusercontent.com/zha0gongz1/html-Absorber/main/pic/1656234842611.jpg">
</center>

<p align="center">保存结果</p>

**注：`-output`为追加式输出到文件output.txt中。**
