# g-snoop
This project provides a demo of a very simple github dashboard. It consists of a web server written in Go that accesses Github's RESTFul API which is documented here:

​	https://developer.github.com/v3/

This project was prepared as a simple proof of concept. The goal is to demonstrate the feasibility to automate access to Github repositories. It currently allows one to perform simple web based operations to create and delete repositories and to display the current list of repositories for a demo account: 

​	https://github.com/g-snoop

Many companies are considering or implementing automate access to github. Hopeful, overtime, we can develop this project to either serve as a useful reference on how to automate access to Github or even potentially evolve into a usably production strength free platform.   

### Live Demo

You can access a live demo running on an AWS t2.micro server:

​	 http://35.165.161.246/

### Installation

```
git clone https://github.com/MarkGisi/g-snoop.git
```

You will need the following despondencies:

```
	go get "github.com/google/go-github/github"
	go get "github.com/gorilla/mux"
	go get "golang.org/x/oauth2"
```

 Execute the build in the main directory:

```
go build
```

You may want to set the port to something other then 80 which is the default in the following config file:

```
g-snoop_config.json
```

Then let it rip:

```
./g-snoop
```

and you will see the follow:

```
Configuration:
-----------------------------------------------
github account   	=  g-snoop
github url      	=  https://github.com/g-snoop
token           	=  57651ab68ac4acd3fce820b6179dddcc19a6e74d
http port       	=  80
debug on          	=  true
verbose on        	= true
config  reload      =  true
running server on port: 80 ...

```

### Feedback & Contributing

You can contact me at mark_gisi@yahoo.com

