all:
	go/8g goweb.go
	go/8l goweb.8

clean:
	rm goweb.8 8.out
