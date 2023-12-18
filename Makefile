.DEFAULT_GOAL := build

# determine whether to use podman or docker for relevant commands
ifeq ($(shell command -v podman 2> /dev/null),)
    DOCKER=docker
else
    DOCKER=podman
endif

# usage: make build
#
# note: this assumes that the "compose" php application is present and daux is
# installed in its default install location, which is ~/.compose/vendor/bin
#
# you can install daux via composer like this: composer global require
# daux/daux.io
#
# see daux documentation here: https://daux.io/Getting_Started.html#install
#
# warning: this may delete the ./docs directory, so make sure you don't have
# anything important in there
build:
	~/.composer/vendor/bin/daux generate \
		--configuration config.json \
		--destination=./docs

# this will run a temporary container that serves the contents of the ./docs
# directory for you to test locally in your browser on port 8099
#
# press ctrl+c to quit
serve:
	@echo ""
	@echo "This will run an nginx container locally that serves ./docs so that you"
	@echo "can verify your changes in a browser."
	@echo ""
	@echo "To exit, press ctrl+c and the container will be destroyed."
	@echo ""
	@echo "Starting container in 5 seconds..."
	@sleep 5
	@echo ""
	-$(DOCKER) rm -f bluetuith-docs
	$(DOCKER) run -p 8099:80 --rm -it --name bluetuith-docs \
		-v $(PWD)/docs:/usr/share/nginx/html:z \
		docker.io/library/nginx:alpine
