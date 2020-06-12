# mdiff

mdiff is **multi tab diff tool on cli, and enable to apply differrece to target file!**

## demo

![demo1](https://github.com/yasutakatou/mdiff/blob/pic/demo.gif)

## solution

when you compare file on terminal , you use diff command often.<br>
diff command usualy is compare two differece of file.
<br>
If you want to compare many file, you use another tool without use terminal. Isn't it?<br>
For example, you copy and paste to your diff  tool  on windows. **Isn't it troublesome?**<br>
this tool is solution for you.<br>

## features

 - compare different on multi tab
 - difference text is colored
 - can search words
 - can apply difference  to target file
 - multi platform support by golang
 - **sjis support**

support linux and windows. and rasbian(arm)!<br>

![demo2](https://github.com/yasutakatou/mdiff/blob/pic/pi.png)

 ## install

*If you want to put it under the path, you can use the following.*

```
go get github.com/yasutakatou/mdiff
```

*If you want to create a binary and copy it yourself, use the following.*

```
git clone https://github.com/yasutakatou/mdiff
cd mdiff
go build mdiff.go
```

or download binary from [release page](https://github.com/yasutakatou/mdiff/releases).<br>
save binary file, copy to entryed execute path directory.<br>

## uninstall

```
delete that binary.
```

del or rm command. *(it's simple!)*

## usecase

You can set filename to compare difference like diff command. You can set any files.<br>

```
mdiff 1.txt 2.txt 3.txt
```
you can compare 1.txt, 2.txt and 3.txt<br>
*note) 1st arg file to be difference master. always compare master to another.*

## operations

**[h]**<br>
display to help of key control.<br>

**[Esc,q]**<br>
exit to program.<br>

**[→,h,Space]**<br>
scroll down.<br>
note) scroll lines to depend your terminal sizes.<br>

**[←,l]**<br>
scroll up.<br>
note) scroll lines to depend your terminal sizes.<br>

**[↑,k]**<br>
one line up.<br>

**[↓,j]**<br>
one line down.<br>

**[Tab,x]**<br>
change to next tab.<br>

**[BS,z]**<br>
back to prev tab.<br>

**[Enter,c]**<br>
commit master file to file of you viewing<br>
when you input "y", copy text to target file of you viewing from you displayed<br>
*note) not viewing part isn't change.*<br>
<br>
when you input "a", all copy master file to target file<br>
*note) this function use to apply all text*<br>
*note) If your text wrote by Shift JIS, this tool convert to utf-8 inside. But, when you commit, this tool return to Shift JIS.*<br>

**[Home:/]**<br>
search to master file by word. input word and press enter to search.<br>
escape key is return to compare difference mode.<br>

## License

MIT
