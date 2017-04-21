package rsync

import (
	"os/exec"
	"fmt"
	"strconv"
	"os"
	"path/filepath"

)

type Rsync struct  {
	cmd *exec.Cmd
	src string
	dest string
	host string
	user string
	pkey string
	port int
	ex string
	ix string
	isok bool
	del bool
	ck bool



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
		user    :u,
		port 	:22,
		isok	:true,

	}
}


func (r *Rsync) setpath(src string,dest string) *Rsync {
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



func (r *Rsync) setType(ty int) *Rsync {
	port:="-p "+ strconv.Itoa(r.port)
	user:=""
	eee:=""
	if len(r.user) !=0 {
		eee="-e"
		user="-l " + r.user
	}
	pkey:=""
	if len(r.pkey) !=0 {
		pkey="-i " + r.pkey
	}

	ex:=""
	if len(r.ex) !=0 {
		ex="--exclude=.git --exclude-from=" + r.ex
	}

	ix:=""
	if len(r.ix) !=0 {
		ix="--files-from=" + r.ix
	}
	src:=r.src
	dest:=r.dest
	if len(r.host) !=0 {
		dest=r.host + ":" + r.dest
	}
	del:=""
	if r.del {
		del="--delete"
	}
	ck:=""
	if r.ck {
		ck="-c"
	}
	rsh:=""
	if len(r.host) !=0 && len(r.user) !=0 {
		rsh=fmt.Sprintf("ssh %s %s %s -o 'StrictHostKeyChecking no'",user,pkey,port)
	}
	args:=[]string{"",eee,rsh,"-avz",del,ix,ex,src,dest,ck}
	for _,v:= range args{
            	if len(v) !=0 {
               		r.cmd.Args= append(r.cmd.Args,v)
            	}
        }
	return r
}

func (r *Rsync) To(){
	fmt.Println(r.cmd.Args)
	if !r.isok { fmt.Println("Not Ready") }
	c := make(chan interface{} ,1024)
	go func(){
		out,err:= r.cmd.CombinedOutput()
		c <-string(out)
		c <-err
	}()
	fmt.Println(<-c)
}
