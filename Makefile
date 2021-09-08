all:
	GOOS=windows GOARCH=386 go build -o build/doge-windows-i386.exe doge/main
	GOOS=windows GOARCH=amd64 go build -o build/doge-windows-x64_86.exe doge/main
	GOOS=linux GOARCH=amd64 go build -o build/doge-linux-x64_86 doge/main
	GOOS=linux GOARCH=386 go build -o build/doge-linux-i386 doge/main
	zip -q build/doge-windows-i386.zip build/doge-windows-i386.exe
	zip -q build/doge-windows-x64_86.zip build/doge-windows-x64_86.exe 
	tar -zcvf build/doge-linux-i386.zip build/doge-linux-i386
	tar -zcvf build/doge-linux-x64_86.zip build/doge-linux-x64_86

clean:
	rm build/doge-*