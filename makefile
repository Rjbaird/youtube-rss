SHELL := /bin/bash
.PHONY=start stop

ifndef VERBOSE
.SILENT:
endif

APP=server

$(APP): *.go
	go build -o $(APP)
	echo "built" $(APP)

.restartServer: *.go start
	touch .restartServer

start: server.PID

server.PID: server
	kill `cat server.PID` 2> /dev/null || echo "Nothing to kill before starting $(APP)"
	./$(APP) & echo $$! > $@; 
	echo "Started $(APP) on with PID `cat server.PID` "

watch: clean server.PID
	while true ; do \
		date; \
		make .restartServer ;\
		fswatch *.go -1 ; \
	done

clean:
	killall $(APP) 2> /dev/null || echo ""
	touch .restartServer server.PID 
	rm .restartServer server.PID $(APP) 2> /dev/null || echo ""

prettier:
	npm run prettier