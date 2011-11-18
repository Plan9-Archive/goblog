all:
	go/8g goweb.go
	go/8l goweb.8

clean:
	rm goweb.8 8.out

install: all
	mkdir -p $home/www/b
	cp page.html archive.html $home/www/b
	cp 8.out $home/www/goweb
