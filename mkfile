all:
	go/8g goweb.go
	go/8l goweb.8

clean:
	rm goweb.8 8.out

copy:
	mkdir -p $home/www/b
	cp page.html archive.html $home/www/b


install: all copy
	cp 8.out $home/www/goweb
