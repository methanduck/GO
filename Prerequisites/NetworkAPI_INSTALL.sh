#! /bin/bash
if [ $(id -u) -ne 0 ]
then 
    echo "Please Run as ROOT(sudo)!!!!"
else
    exist = $(dpkg -l | grep golang)
    if test -z "$exist"
    then
        echo "GOLANG INSTALLATION INITIATED"
        apt-get install golang-go
        if test -n "$(dpkg -l | grep golang)" 
        then
            echo "GO : $(go version)"
            echo "PLEASE CONFIG GOPATH (IF BLANK : /home/go) : "
            read input
            if test -z "$input"
            then
                mkdir $HOME/go
                export GOPATH=$HOME/go
                echo "export GOPATH=$HOME/go" >> $HOME/.bashrc
            else
                mkdir $HOME/go
                export GOPATH=$input
                echo "export GOPATH=$input" >> $HOME/.bashrc
            fi
            echo "GOLANG INSTALLATION COMPLETED"
        else
        echo "GOLANG INSTALLATION FAILED"
        fi  

        go get github.com/methanduck/GO/InteractiveSocket
        cd $GOPATH/src/github.com/methanduck/GO/InteractiveSocket/main
        go build main.go

        echo "RUN PROGRAM? [y/n]"
        read YesorNo
        if test -z "$YesorNo"
        then
            $GOPATH/src/github.com/methanduck/GO/InteractiveSocket/main/main
        fi
    fi
fi
