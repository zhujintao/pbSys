package rsync

import (
	"os/exec"
	"fmt"
	"strconv"
	"os"
	"path/filepath"
	"golang.org/x/net/websocket"

	"bufio"
	"github.com/gin-gonic/gin"
)

type Rsync struct  {

	cmd *exec.Cmd
	src string
	dest string
	host string
	User string
	Pkey string
	Port int
	Ex string
	Ix string
	isok bool
	del string
	ck string




}

func NewRsync(args ...string) *Rsync{

	h:=""
	u:=""
	if len(args) == 1{
		h=string(args[0])
	}
	if len(args) == 2 {
		h=string(args[0])
		u=string(args[1])
	}

	return &Rsync{
		cmd	:exec.Command("/usr/bin/rsync"),
		host	:h,
		User    :u,
		Port 	:22,
		isok	:true,

	}
}


func (r *Rsync) Setpath(src string,dest string) *Rsync {

	f,err :=os.Stat(src)
	if !os.IsExist(err) && err != nil {
		r.isok=false
		return r
	}
	if f.IsDir(){
		p:= fmt.Sprintf("%c",os.PathSeparator)
		src =filepath.Join(src) + p
	}
	r.src 	= src
	r.dest	= dest
	return r

}



func (r *Rsync) SetType() *Rsync {




	port:="-p "+ strconv.Itoa(r.Port)
	user:=""
	eee:=""
	if len(r.User) !=0 {
		eee="-e"
		user="-l " + r.User
	}
	pkey:=""
	if len(r.Pkey) !=0 {
		pkey="-i " + r.Pkey
	}

	ex:=""
	if len(r.Ex) !=0 {
		ex="--exclude-from=" + r.Ex
	}

	ix:=""
	if len(r.Ix) !=0 {
		ix="--files-from=" + r.Ix
	}
	src:=r.src
	dest:=r.dest
	if len(r.host) !=0 {
		dest=r.host + ":" + r.dest
	}
	del:=""
	if len(r.del) !=0 {
		del="--delete"
	}
	ck:=""
	if len(r.ck) !=0 {
		ck="-c"
	}
	rsh:=""
	if len(r.host) !=0 && len(r.User) !=0 {
		rsh=fmt.Sprintf("ssh %s %s %s -o 'StrictHostKeyChecking=no'",user,pkey,port)
	}
	args:=[]string{"",eee,rsh,"-avcz","--exclude=.git",del,ix,ex,src,dest,ck}
	for _,v:= range args{
            	if len(v) !=0 {
               		r.cmd.Args= append(r.cmd.Args,v)
            	}
        }
	return r
}

func (r *Rsync) To(c chan int,i int,n string,ws []*websocket.Conn){
	fmt.Println(r.cmd.Args)
	if r.isok {

		//out,err:= r.cmd.CombinedOutput()
		stdout,_:=r.cmd.StdoutPipe()
		r.cmd.Start()
		r.cmd.Run()
		defer r.cmd.Wait()
		b:=bufio.NewScanner(stdout)

		for b.Scan() {

			for _,w:=range ws {
				websocket.JSON.Send(w,gin.H{
					"id":i,
					"item":n,
					"message":b.Text(),
				})
			}
		}

		for _,w:=range ws {

			websocket.JSON.Send(w,gin.H{
					"id":i,
					"item":n,
					"message":"done:true",
				})
		}

		c <- i



	}else{


		for _,w:=range ws {

			websocket.JSON.Send(w, gin.H{
				"id":i,
				"item":n,
				"message":" Failed: configuration error is not updated",
			})
		}


	}



}

