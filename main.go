package main

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"pbsys/rsync"

	"io"
	"net/url"

	"github.com/gin-gonic/contrib/sessions"

	"fmt"

	"net/http"
	"io/ioutil"
	"encoding/json"

	"strconv"
	"path/filepath"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/websocket"

	"os/exec"
	"strings"


	"regexp"

	"os"
)



type Project struct {
	Id      int
	Name    string
	Url     string `json:"ssh_url_to_repo"`
	Message  []string
	Changes []Changes
	Title   string
}
type Msg struct {
	Message string
	Branch_name string

}
type Changes struct {
	Diff string

}

type Groups struct {
	Id int
	Name string
	Projects []Project
}



type User struct {
	Name string
	Username string
	Email string
	Token string `json:"private_token"`
	Message string
	Role string

}

type Reco struct {
	Name string
	Branch string
	Branch_test string
	Branch_test_status string
	Server string
	Taskid string
	Owen	string
}

type Pbsrv struct {
	Id int
	Name string
	Type string
	IP string
	Port int
	User string
	Path string
	Srvid string
	Repo string

}




var wsclients []*websocket.Conn
var usertk map[string]string

var repoinfos  map[int]map[string]string
var srvinfos map[int]map[string]string

var iteminfos map[int]map[string]string
var db *sql.DB

var store = sessions.NewCookieStore([]byte("secret1"))




func InitDB(filepath string) *sql.DB {

	db,err := sql.Open("sqlite3",filepath)
	if err != nil {panic(err)}
	if db == nil {panic("db nil")}


	sql_table:=`
CREATE TABLE IF NOT EXISTS [server] (
  [Name] TEXT,
  [Type] INT,
  [Ip] TEXT,
  [Port] INT,
  [User] TEXT,
  [Path] TEXT,
  [Repo] TEXT);
	`

	sql_table2:=`
CREATE TABLE IF NOT EXISTS [item] (
  [Name] TEXT,
  [Srvid] TEXT,
  [Path] TEXT);
	`

	sql_table3:=`
CREATE TABLE IF NOT EXISTS [reco] (
  [repo_id] TEXT,
  [repo_name] TEXT,
  [branch] TEXT,
  [branch_test] TEXT,
  [branch_test_status] TEXT,
  [branch_prod] TEXT,
  [tester] TEXT,
  [ready ] TEXT,
  [accept] TEXT,
  [srv_id] CHAR);
  	`
	_,err = db.Exec(sql_table)
	if err != nil {panic(err)}

	_,err = db.Exec(sql_table2)
	if err != nil {panic(err)}

	_,err = db.Exec(sql_table3)
	if err != nil {panic(err)}
	return  db
}


func getpjSrv(ty int) []Pbsrv {
	sql_readall := `
	SELECT rowid,Name,Type,Ip,Port,User,Path,Repo  FROM server where type=?
	`
	rows, err := db.Query(sql_readall,ty)
	if err != nil { panic(err) }
	defer rows.Close()

	var result []Pbsrv
	for rows.Next() {
		item := Pbsrv{}
		err2 := rows.Scan(&item.Id, &item.Name,&item.Type,&item.IP,&item.Port,&item.User,&item.Path,&item.Repo)
		if err2 != nil { panic(err2) }
		result = append(result, item)
	}
	return result
}


func getpjItem(db *sql.DB) []Pbsrv {
	sql_readall := `
	SELECT rowid,Name,Path,Srvid FROM item
	`
	rows, err := db.Query(sql_readall)
	if err != nil { panic(err) }
	defer rows.Close()

	var result []Pbsrv
	for rows.Next() {
		item := Pbsrv{}
		err2 := rows.Scan(&item.Id,&item.Name,&item.Path,&item.Srvid)
		if err2 != nil { panic(err2) }
		result = append(result, item)
	}
	return result
}


func getallPj() []Project {

	url,_ :=url.Parse("http://repo.ab.pub/api/v3/projects/all?page=&per_page=1000")
	token := url.Query()
	token.Set("private_token","ab.pub")
	url.RawQuery=token.Encode()

	r,_:=http.Get(url.String())
	defer r.Body.Close()
	rData,_ := ioutil.ReadAll(r.Body)
	var projects []Project
	json.Unmarshal([]byte(rData),&projects)
	return projects
}


func getownPj(tk string) []Project {

	url,_ :=url.Parse("http://repo.ab.pub/api/v3/projects?page=&per_page=1000")

	token := url.Query()
	token.Set("private_token",tk)
	url.RawQuery=token.Encode()

	r,_:=http.Get(url.String())
	defer r.Body.Close()
	rData,_ := ioutil.ReadAll(r.Body)
	var projects []Project
	json.Unmarshal([]byte(rData),&projects)

	return projects
}


func getPjBranchs(id,tk string) []Project {

	url,_ :=url.Parse(fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/repository/branches",id))
	token := url.Query()
	token.Set("private_token",tk)
	url.RawQuery=token.Encode()

	r,_:=http.Get(url.String())
	defer r.Body.Close()
	rData,_ := ioutil.ReadAll(r.Body)
	var project []Project
	json.Unmarshal([]byte(rData),&project)


	return project

}

func regbuild(exp,chr string) string {
	r,_:=regexp.Compile(exp)
        d,_:=regexp.Compile(`\D+`)
	istr:=strings.Join(r.FindAllString(chr,-1),"")

	return "b"+strings.Join(d.Split(istr,-1),"")
}


func crePjBranch(id,ref,tk,owen string) (string,bool) {

	n,_:=regexp.Compile("test_b[0-9]+_")
	d,_:=regexp.Compile("^dev_fix_|^dev_")

	dev_suffix:=strings.Join(d.Split(ref,-1),"")
	nmax:=0
	for _,b:=range getPjBranchs(id,tk) {
		test_suffix:=strings.Join(n.Split(b.Name,-1),"")

		if dev_suffix == test_suffix {
			istr:=strings.Join(strings.Split(strings.Join(n.FindAllString(b.Name,-1),""),"test_b"),"")
			i,_:=strconv.Atoi(string([]rune(istr)[:len(istr)-1]))

			if i > nmax {
				nmax=i
			}
		}
	}

	var test_status string
	db.QueryRow("select branch_test_status from reco where repo_id=? and branch=? and branch_test=? and owen=?",id,ref,nmax,owen).Scan(&test_status)

	if test_status == "pass" || nmax == 0 {

		branch_name:= "test_b" + fmt.Sprintf("%.3d",nmax+1) + "_" + dev_suffix
		mtitle := "[develop->" + ref + "]->" + branch_name

		msg,pullok:=pullUpdata(id,mtitle,"develop",ref,tk)

		if pullok {

			r,_ :=http.PostForm(fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/repository/branches",id),url.Values{"branch_name":{branch_name},"ref":{ref},"private_token":{tk}})
                        defer r.Body.Close()
                        rData,_ := ioutil.ReadAll(r.Body)
                        var project Project
                        json.Unmarshal([]byte(rData),&project)

			if len(project.Name) !=0 {
				stmt,_:=db.Prepare("insert into reco(repo_id,branch,branch_test,branch_test_status,owen) values(?,?,?,?,?)")

				stmt.Exec(id,ref,fmt.Sprintf("b%.3d",nmax+1),"queue",owen)

			}

                        return fmt.Sprintf("b%.3d",nmax+1),true

		}
		return msg,false

	}

	switch test_status {

		case "queue": return ref + " queue already exists",false
		case "test" : return ref + " testing...",false

	default:
		return "not allow.",false
	}
}


func delPjBranch(id,b,tk string) bool {
	client:=&http.Client{}

	delURL:=fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/repository/branches/%s",id,b)
	reqDl, _ := http.NewRequest("DELETE", delURL, nil)
	reqDl.Header.Set("private_token",tk)
	clientDL,_:=client.Do(reqDl)
	defer clientDL.Body.Close()


	rData,_ := ioutil.ReadAll(clientDL.Body)
	var msg Msg
	json.Unmarshal([]byte(rData),&msg)
	if msg.Branch_name == b {
		return true
	}
	return false
}


func pullUpdata(id,title,s,t,tk string) (string,bool) {

	client:=&http.Client{}
	clientCmr,_ :=http.PostForm(fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/merge_requests",id),url.Values{"title":{title},"source_branch":{s},"target_branch":{t},"private_token":{tk}})
	defer clientCmr.Body.Close()
	cmrData,_ := ioutil.ReadAll(clientCmr.Body)
	var respCmr Project
	json.Unmarshal([]byte(cmrData),&respCmr)
	mid := respCmr.Id
	if strings.Contains(strings.Join(respCmr.Message,""),"Cannot Create: This merge request already exists") {
		reg,_:=regexp.Compile(`\[".*."\]`)
		ntitle:=strings.Join(strings.Split(strings.Join(strings.Split(reg.FindString(respCmr.Message[0]),`["`),""),`"]`),"")
		getmrURL,_ :=url.Parse(fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/merge_requests",id))
		token := getmrURL.Query()
		token.Set("private_token",tk)
		token.Add("state","opened")
		getmrURL.RawQuery=token.Encode()
		clientGmr,_:=http.Get(getmrURL.String())
		defer clientGmr.Body.Close()
		gmrData,_ := ioutil.ReadAll(clientGmr.Body)
		var respGmrs []Project
		json.Unmarshal([]byte(gmrData),&respGmrs)
		for _,v:=range respGmrs{
			if v.Title == ntitle {
				mid=v.Id
			}
		}
	}

	changeUrl,_ :=url.Parse(fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/merge_request/%d/changes",id,mid))
	token := changeUrl.Query()
	token.Set("private_token",tk)
	changeUrl.RawQuery=token.Encode()
	clientCh,_:=http.Get(changeUrl.String())
	defer clientCh.Body.Close()
	chData,_ := ioutil.ReadAll(clientCh.Body)
	var respCh Project
	json.Unmarshal([]byte(chData),&respCh)
	if len(respCh.Changes) !=0 {

		mergeURL := fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/merge_request/%d/merge", id, mid)
		reqMr, _ := http.NewRequest("PUT", mergeURL, nil)
		reqMr.Header.Set("private_token",tk)
		clientMr,_:=client.Do(reqMr)
		defer clientMr.Body.Close()
		meData,_ := ioutil.ReadAll(clientMr.Body)
		var respMr Msg
		json.Unmarshal([]byte(meData),&respMr)
		if len(respMr.Message) != 0 {
			if strings.Contains(respMr.Message, "Branch cannot be merged") {
				return s + "->" + t + " conflicts\n\n"+ respCh.Changes[0].Diff      ,false
			}
		}else {
			return s + "->" + t + " merged ok",true
		}

	}else {

		upmrURL := fmt.Sprintf("http://repo.ab.pub/api/v3/projects/%s/merge_request/%d?state_event=close", id, mid)
		req, _ := http.NewRequest("PUT", upmrURL, nil)
		req.Header.Set("private_token", tk)
		clientUmr, _ := client.Do(req)
		defer clientUmr.Body.Close()

	}
	return "The changes were not merged into "+t,true
}


func gitlabAuth(u,p string) User  {

	r,_ := http.PostForm("http://repo.ab.pub/api/v3/session",url.Values{"login":{u},"password":{p}})
	defer r.Body.Close()
	rData,_ := ioutil.ReadAll(r.Body)
	var user User
	json.Unmarshal([]byte(rData),&user)
	user.Role="deve"
	return user

}


func localAuth(u,p string) User {


	//hp,_:=bcrypt.GenerateFromPassword([]byte(p),bcrypt.DefaultCost)
	var user User
	var password string
	db.QueryRow("select password,name,username,role,email from user  where username=?",u).Scan(&password,&user.Name,&user.Username,&user.Role,&user.Email)
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(p))
	if err != nil {
		user=User{}
		user.Message="password auth err"
	}
	return user

}


func git(gitdir,cmd string) string {

        c:="--git-dir="+filepath.Join(gitdir,".git")
        if strings.Split(cmd," ")[0] != "clone" {
                c+=" " + "--work-tree=" + filepath.Join(gitdir)
        }

        c+=" " + cmd
        out,err:=exec.Command("git",strings.Split(c," ")...).CombinedOutput()

        if err != nil   {
                return string(out)
        }
        return ""
}



func loginCheck(c *gin.Context) User {

	session := sessions.Default(c)
	curr_user:=session.Get("curr_user")
	if curr_user != nil {
		var u User
		json.Unmarshal([]byte(curr_user.(string)),&u)
		return u
	}
	c.Redirect(302,"/")
	return User{}

}
func loginCheckWs(ws *websocket.Conn) User {
	session,_:=store.Get(ws.Request(),"mysessiosdfasdfn")
	curr_user:=session.Values["curr_user"]
	if curr_user != nil {
		var u User
		json.Unmarshal([]byte(curr_user.(string)),&u)
		return u
	}
	ws.Close()
	return User{}
}


func login(c *gin.Context){
	session := sessions.Default(c)
	if c.Request.Method == "GET" {
		c.HTML(200, "login.html", gin.H{})
	}

	if c.Request.Method == "POST" {
		u:=User{Message:"plass login"}

		switch c.PostForm("role") {
		case "test","prod":
			u=localAuth(c.PostForm("username"),c.PostForm("password"))

		case "deve":
			u=gitlabAuth(c.PostForm("username"),c.PostForm("password"))
			if len(usertk) == 0 {
				usertk= make(map[string]string)
			}
			usertk[u.Username]=u.Token
		}

		if len(u.Message) == 0 {
			b,_:=json.Marshal(u)
			session.Set("curr_user",string(b))
			err:=session.Save()
			if err != nil {
				fmt.Println("session save error",err)
			}
			switch u.Role {
			case "test"	:c.Redirect(302,"/upto/testStatus")
			}

			c.Redirect(302,"/upto")

		}else{
			c.Redirect(302,"/")
		}
	}
}







func listReco(ty string,u User) (recos []Reco) {

	var reco Reco

	if ty == "owen" {
		rows, _ := db.Query("select a.repo_name,a.branch,a.branch_test,a.branch_test_status,a.rowid,b.name from reco a,server b where a.srv_id = b.rowid and a.owen=?", u.Username)
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&reco.Name, &reco.Branch, &reco.Branch_test, &reco.Branch_test_status, &reco.Taskid, &reco.Server)
			recos = append(recos, reco)

		}
		return recos
	}

	rows, _ := db.Query("select a.repo_name,a.branch,a.branch_test,a.branch_test_status,a.rowid,a.owen,b.name from reco a,server b where a.srv_id = b.rowid and a.branch_test_status !='queue'")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&reco.Name, &reco.Branch, &reco.Branch_test, &reco.Branch_test_status, &reco.Taskid,&reco.Owen, &reco.Server)
		recos = append(recos, reco)
	}

	return recos

}

func addReco(pjid,branch,srvid string,u User,ws *websocket.Conn){
	if pjid == "-1" {return}
	var npass string
	db.QueryRow("select count(branch) from reco where repo_id=? and srv_id=? and branch_test_status != 'pass'",pjid,srvid).Scan(&npass)
	inpass,_:=strconv.Atoi(npass)
	if inpass != 0 {

		websocket.JSON.Send(ws,gin.H{"tag":"message","body":"server not allow add other branch for test"})
		return

	}

	tmesg,err:=crePjBranch(pjid,branch,u.Token,u.Username)
	if !err {
		websocket.JSON.Send(ws,gin.H{"tag":"message","body":tmesg})
		return
	}

	stmt,_:=db.Prepare("update reco set srv_id=?, repo_name=? where  repo_id = ? and branch = ?")
	id,_:=strconv.Atoi(pjid)
	stmt.Exec(srvid,repoinfos[id]["name"],pjid,branch)
	var server,taskid string
	db.QueryRow("select name from server where rowid = ?",srvid).Scan(&server)
	db.QueryRow("select rowid from reco where repo_id=? and branch=? and branch_test=?",pjid,branch,tmesg).Scan(&taskid)

	websocket.JSON.Send(ws,gin.H{
		"tag"	:"add_reco",
		"body"	:gin.H{
			"pjname":repoinfos[id]["name"],
			"branch":branch,
			"branch_test":tmesg,
			"server":server,
			"taskid":taskid,

		},
	})

}


func delReco(body string,u User,ws *websocket.Conn){
		var recos []Reco
		var reco Reco
		rows,_:=db.Query("select a.repo_name,a.branch,a.branch_test,b.name from reco a,server b where a.srv_id = b.rowid and a.owen=?",u.Username)
		defer rows.Close()
		for rows.Next() {

			rows.Scan(&reco.Name, &reco.Branch, &reco.Branch_test, &reco.Server)
			recos = append(recos, reco)
		}

		for _,v := range recos {
			p:=fmt.Sprintf("%s%s%s%sx",v.Name,v.Branch,v.Branch_test,v.Server)
			//p:=fmt.Sprintf("%s%s%s%s%s%sx",v.Name,v.Branch,strings.Replace(v.Ifile,"\n","",-1),v.Branch_test,v.Server,v.Path)

			if p == body {


				var srv_id, dpid string

				db.QueryRow("select rowid from server where name=?",v.Server).Scan(&srv_id)
				db.QueryRow("select repo_id from reco where repo_name=? and srv_id=? and owen=?",v.Name,srv_id,u.Username).Scan(&dpid)

				d,_:=regexp.Compile("^dev_fix_|^dev_")
				dev_suffix:=strings.Join(d.Split(v.Branch,-1),"")

				if delPjBranch(dpid,"test_"+ v.Branch_test + "_" + dev_suffix,u.Token) {
					stmt,_  := db.Prepare("delete from reco where repo_name = ? and branch=? and branch_test = ? and branch_test_status=? and srv_id=? and owen=?")
					stmt.Exec(v.Name,v.Branch,v.Branch_test,"queue",srv_id,u.Username)
					websocket.JSON.Send(ws,gin.H{"tag":"message","body":"del success!"})
				}
			}
		}
}


func listBranchs(pjid string,u User,ws *websocket.Conn){

	var branchs []map[string]string
	for _,v:= range getPjBranchs(pjid,u.Token){

		item := make(map[string]string)

		if strings.Contains(v.Name,"dev_") {

			item["id"] = v.Name; item["text"] = v.Name
			branchs = append(branchs, item)
		}
	}
	websocket.JSON.Send(ws,gin.H{"tag":"list_branchs","body":branchs})

}


func listServers(pjname string,u User,ws *websocket.Conn){

//	var tySrv map[string]interface{}
	var servers []map[string]string
	ty := 2
	if u.Role == "prod" { ty=3 }
	for _,v:= range getpjSrv(ty) {
		item := make(map[string]string)
		for _,r:= range strings.Split(v.Repo,"|"){
			if pjname == r {
				item["id"]=strconv.Itoa(v.Id)
				item["text"]=v.Name
				servers=append(servers,item)
			}
		}

	}

	websocket.JSON.Send(ws,gin.H{"tag":"list_servers","body":servers})

}



func modRecoStatus(status string,u User,wss []*websocket.Conn){

	if status == "queue" {
		rows, _ := db.Query("select a.rowid,a.repo_name,a.branch,a.branch_test,a.branch_test_status,a.owen,b.name from reco a,server b where branch_test_status='queue' and a.srv_id=b.rowid and owen=?", u.Username)
		defer rows.Close()
		for rows.Next() {
			var taskid,repo_name,branch,branch_test,branch_test_status,owen,server string
			rows.Scan(&taskid,&repo_name,&branch,&branch_test,&branch_test_status,&owen,&server)

			for _,ws:= range wss {
				websocket.JSON.Send(ws, gin.H{"tag":"mod_status_q", "body":gin.H{"taskid":taskid,
					"pjname":repo_name,
					"branch":branch,
					"branch_test":branch_test,
					"branch_test_status":"wait",
					"owen":owen,
					"server":server},
				})
			}

		}

		stmt, _ := db.Prepare("update reco set branch_test_status ='wait' where branch_test_status='queue' and owen=?")
		stmt.Exec(u.Username)

	}

}


func receTest(body string,wss []*websocket.Conn){


	var dat map[string]interface{}
	var status string
	json.Unmarshal([]byte(body),&dat)

	if dat["value"] == "wait->Ok"{


		var server,ip,port,user,pkey,path,pjid,repo,branch,build string
		db.QueryRow("select a.name,a.ip,a.port,a.user,a.pkey,a.Path,b.repo_id,b.repo_name,b.branch,b.branch_test from server a ,reco b where b.srv_id = a.rowid and b.rowid =?",dat["id"]).Scan(&server,&ip,&port,&user,&pkey,&path,&pjid,&repo,&branch,&build)

		d,_:=regexp.Compile("^dev_fix_|^dev_")
		dev_suffix:=strings.Join(d.Split(branch,-1),"")
		branch_name:= "test_"+ build  + "_" + dev_suffix

		ddir:=filepath.Join(path,repo)
		gitdir:=filepath.Join("tmp",filepath.Join(server,repo))
		ipjid,_:=strconv.Atoi(pjid)
		clone:=fmt.Sprintf("clone -b %s %s %s",branch_name,repoinfos[ipjid]["url"],gitdir)
		result:=git(gitdir,clone)

		if strings.Contains(result,"fatal: The remote end hung up unexpectedly"){
				fmt.Println("Error: Deploy Keys (GitLab)")
				return
		}


		if strings.Contains(result,"already exists") {

			if strings.Contains(git(gitdir,"fetch"),"fatal: Not a git repository") {
				os.RemoveAll(gitdir)
				result=git(gitdir,clone)
		}
			result=git(gitdir,"checkout -f "+ branch_name)
			result=git(gitdir,"pull")
			fmt.Println(result)
		}


		if len(result) == 0 {

			r:=rsync.NewRsync(ip,user)
			r.User=user
			r.Pkey=pkey


			exf:=filepath.Join(gitdir,".git","conflist")
			if _,err :=os.Stat(exf); err == nil {
				r.Ex=exf
			}
			r.Setpath(gitdir,ddir)
			//status="testing"
			stmt,_:=db.Prepare("update reco set branch_test_status='testing' where rowid=?")
			stmt.Exec(dat["id"])
			status="testing"

		}else {

			status="deploy error!"

		}

		//ioutil.WriteFile(filepath.Join(gitdir,".git","ixftmp"),[]byte(ifile),0644)




	}


	if dat["value"] == "testing->Ok"{

		stmt,_:=db.Prepare("update reco set branch_test_status='pass' where rowid=?")
		stmt.Exec(dat["id"])
		status="pass"
		var owen string
		db.QueryRow("select owen from reco where rowid=?",dat["id"]).Scan(&owen)
		fmt.Println(usertk[owen])


	}

	if len(status) !=0 {

		for _,ws:= range wsClients {
			websocket.JSON.Send(ws, gin.H{"tag":"receive_test", "body":gin.H{"id":dat["id"], "value":status}})
		}
	}




}




func requUpto(pjid string, u User){

	

}


var wsClients []*websocket.Conn
type MsgJson struct {

	Tag string
	Body string
	Extra1 string `json:"ty"`
	Extra2 string `json:"pjid"`
	Extra3 string `json:"branch"`
	Extra4 string `json:"srvid"`

}
func active(ws *websocket.Conn){
	u:=loginCheckWs(ws)
	defer ws.Close()
	wsClients=append(wsClients,ws)


	var pbjs []map[string]string
	for _,v := range getownPj(u.Token){
		pbj := make(map[string]string)
		pbj["id"]= strconv.Itoa(v.Id)
		pbj["text"]=v.Name
		pbjs=append(pbjs,pbj)
		if len(repoinfos) == 0 {
			repoinfos = make(map[int]map[string]string)
		}
		repoinfos[v.Id]=map[string]string{"url":v.Url,"name":v.Name}
	}

	websocket.JSON.Send(ws,gin.H{"tag":"pbjs","body":pbjs})

	for {
		var msg MsgJson
		err:=websocket.JSON.Receive(ws,&msg)
		if err == io.EOF {
			for i,v := range wsclients {
				if v == ws {
					 wsclients = append(wsclients[0:i],wsclients[i+1:]...)
				}
			}
			return
		}

		switch msg.Tag {

		case "del_reco" 	: delReco(msg.Body,u,ws)
		case "list_branchs" 	: listBranchs(msg.Body,u,ws)
		case "list_servers"	: listServers(msg.Body,u,ws)
		case "add_reco"	: addReco(msg.Extra2,msg.Extra3,msg.Extra4,u,ws)
		case "mod_status"	: modRecoStatus(msg.Body,u,wsClients)
		case "receive_test"	: receTest(msg.Body,wsClients)
		case "requ_upto"	: requUpto(msg.Body,u)

		}
	}
}


func upto(c *gin.Context) {

	u:=loginCheck(c)
	switch c.Param("tyro"){
	case "/testStatus":
		//if u.Role == "deve" { c.Redirect(302,"/upto") }
		if u.Role=="test"{
			c.HTML(200,"reco.html",gin.H{"queue":listReco("all",u),"curr_user":u})
			return
		}
	case "/pb":
		if u.Role=="test"{
			c.HTML(200,"pb.html",gin.H{"queue":listReco("all",u),"curr_user":u})
			return
		}

	}

	if u.Role == "deve" { c.HTML(200, "index.html", gin.H{"queues":listReco("owen", u), "works":listReco("all", u), "curr_user":u}) }
	if u.Role == "test" { c.Redirect(302,"/upto/testStatus") }
	if u.Role == "prod" { c.Redirect(302,"/upto/pb") }
}



func main() {


	db=InitDB("repo.db")
	defer db.Close()
	router := gin.New()

	store.Options(sessions.Options{MaxAge:0})
	router.Use(sessions.Sessions("mysessiosdfasdfn",store))
	router.Static("/libs","libs")
	router.LoadHTMLGlob("templates/*")
	router.GET("/active",gin.WrapH(websocket.Handler(active)))

	router.GET("/",login)
	router.POST("/",login)
	router.GET("/upto",upto)
	router.GET("/upto/*tyro",upto)

	router.Run(":3003")

}

