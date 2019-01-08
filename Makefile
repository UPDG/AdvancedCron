all:
	@./build.sh
build:
	@./build.sh all
clean:
	rm -rf build
install: all
	cp rr /usr/local/bin/acron
uninstall: 
	rm -f /usr/local/bin/acron
