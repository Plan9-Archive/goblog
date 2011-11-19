all: goweb.go
	go/8g goweb.go
	go/8l goweb.8

clean:
	rm goweb.8 8.out

setupblog:
	mkdir -p $home/www/b
	cp page.html archive.html $home/www/b

copyconfig:
	cp config.json /lib/goweb.config

install:
	cp 8.out $home/bin/386/goweb

