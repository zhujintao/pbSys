
from paramiko import SSHClient ,AutoAddPolicy ,RSAKey
from StringIO import StringIO
from subprocess import Popen,PIPE
from shutil import rmtree
from sys import argv
import os,json,ConfigParser,re,getpass
import tornado.web
import tornado.escape
from tornado import gen
from tornado.options import define, options, parse_command_line
from gitlab import Gitlab
from gittle import Gittle
from gittle.auth import GittleAuth





repo_url={}
if os.path.exists('/etc/pbsys/conf.ini'):CONF='/etc/pbsys/conf.ini'
else:CONF='conf.ini'
class CappConfig(ConfigParser.ConfigParser):
    def __init__(self,filename):
        ConfigParser.ConfigParser.__init__(self)
        self.filename=filename
        self.read(filename)
    def optionxform(self,optionstr):
        return optionstr


class sshto():
    def __init__(self,host,port,username,password,key,key_obj=False):
        try:
            if key_obj:
                mykey=RSAKey.from_private_key(StringIO(key))
            else:
                mykey = RSAKey.from_private_key_file(key)
            self.client =SSHClient()
            self.client.set_missing_host_key_policy(AutoAddPolicy())
            self.client.connect(hostname=host,port=port,username=username,password=password,pkey=mykey)
        except:
            self.client=None

    def exe_cmd(self,cmd):
        if self.client:
            _,stdout,stderr = self.client.exec_command(cmd)
            return stderr.read(),stdout.read()
        else:
            return 'err',None

    def exe_cmdOne(self,cmd):
        if self.client:
            _,stdout,stderr = self.client.exec_command(cmd)
            err,out = stderr.read(),stdout.read()
            self.client.close
            return err,out
        else:
            return 'err',None

    def close(self):
        if self.client:
            self.client.close()

def rsync(src,dest,rsync=None,arg=None,pkey=None,check=False):
    if rsync is None :rsync='rsync'
    if arg is None : arg=''
    if pkey:pkey='-i %s'% pkey
    else:pkey=''
    if check:check="-n"
    else:check=''
    #rsync -e "ssh -i /home/git/repositories/rsawww -o 'StrictHostKeyChecking no'" -avz --delete repo/ www@127.0.0.1:/tmp/testing

    rsync_cmd='''%s -e "ssh %s -o 'StrictHostKeyChecking no'" --delete -avcz %s %s %s %s'''%(rsync,pkey,arg,src,dest,check)

    rsync_out=Popen(rsync_cmd,shell=True,stdout=PIPE,stderr=PIPE)
    out,err = rsync_out.communicate()

    return err,out


class BaseHandler(tornado.web.RequestHandler):
    def get_current_user(self):
        user_json = self.get_secure_cookie("id_rsa_user")
        if not user_json: return None
        return tornado.escape.json_decode(user_json)

class AuthLoginHandler(BaseHandler):
    def get(self):
        self.render("login.html")
    def post(self):
        id_rsa = self.request.files.get('id_rsa')
        if id_rsa:id_rsa=id_rsa[0]
        else:self.redirect("/pbsys");return
        parser=CappConfig(CONF)
        _,user=sshto('%s'%parser.get('GITS','address'),int(parser.get('GITS','sshport')),'git','',id_rsa.get('body'),True).exe_cmdOne('auth')
        if user:
            self.set_secure_cookie('id_rsa_user',tornado.escape.json_encode(json.loads(user)),expires_days=None)
            self.redirect("/pbsys")
            return
        else:
            self.redirect("/pbsys/auth/login")


class AuthLogoutHandler(BaseHandler):
    @gen.coroutine
    def get(self):

        self.clear_cookie("id_rsa_user")
        #self.write("You are now logged out")
        self.redirect("/pbsys/auth/login")


class MainHandler(BaseHandler):
    @tornado.web.authenticated

    def get(self):
        projets=get_gitbranch(self.get_current_user()['username'])

        self.render("index.html",sInfo=get_srvlist(),blists=projets)
        #self.render("index.html")







def get_srvlist(sid=None):


    dst={}
    sList=[]
    parser=CappConfig(CONF)
    try:
        parser.items('REPO')
    except:return dst
    for s in parser.items('REPO'):

        suffix='.'+''.join(s[0].split('.')[-1:])
        sl=s[0].split(suffix)[0]
        sList.append(sl)
    slist=list(set(sList))
    for i in range(0,len(slist)):
        dst[i]=slist[i]

    if sid:
        return dst[int(sid)]

    return dst


def get_gitbranch(auth_user):
    blists=[]
    parser=CappConfig(CONF)
    workdirs=parser.get('PATH','workdirs')
    verconf=CappConfig(os.path.join(workdirs,'vercnf'))

    try:
        gl=Gitlab('http://%s'%parser.get('GITS','address') ,'%s'%parser.get('GITS','addretk'))
        gl.auth()
    except:
        return  blists

    for project in gl.projects.list(page=0,per_page=1000):

        bdict={};blist=[];mlist=[]
        for branch in  gl.project_branches.list(project_id=project.id):
            if not re.match(parser.get('GITS','brannot').replace(',','|'),branch.name):
                blist.append(branch.name)
        bdict['bname']=blist
        bdict['pname']=project.name
        bdict['pnumb']=''

        repo_url[project.name]=project.ssh_url_to_repo
        #bdict['sshurl']=project.ssh_url_to_repo
        flag=False

        for key,val in  get_srvlist().viewitems():
             if project.name in val.split('-',1):
                auth_perm=parser.get('REPO','%s.permissi'%val).replace('|c','')
                flag=True
                if len(val.split('-',1)) == 2:

                   if auth_user in  auth_perm.split(','):
                       try:
                           mvers=verconf.get('REPO','%s.versions'%val)
                       except:
                           mvers=''
                       mlist.append({'mnumb':key,'mname':val.split('-')[0],'mvers':mvers})
                      #print mlist
                else:
                   if auth_user in  auth_perm.split(','):
                      bdict['pnumb']=key
                   else:bdict['pnumb']=''

        if mlist: bdict['mlist']=mlist
        else : bdict['mlist']=''

        if not bdict['pnumb']  and not bdict['mlist']:pass
        else:blists.append(bdict)

    return blists



class Publish(BaseHandler):
    @tornado.web.authenticated
    def post(self):
        #authuserList=['jino']
        #if not  self.get_current_user() in authuserList:
        #    self.write('unauthorized users')
        #    return

        parser=CappConfig(CONF)
        self.set_header('Content-Type', 'text/plain')
        #sid=self.request.arguments['sid']


        cid=self.get_argument('cid',None)
        if not cid: self.write('ERROR:invalid branch');return
        sid = self.get_argument('sid', None)
        if not sid: self.write('ERROR:invalid server');return
        dst = get_srvlist(sid)
        servname=parser.get('REPO','%s.servname'%dst)
        host=parser.get('SERVER','%s.hostname'%servname)
        user=parser.get('SERVER','%s.username'%servname)
        pwds=parser.get('SERVER','%s.password'%servname)
        keyd=parser.get('PATH','keysdire')
        pkey=os.path.join(keyd,parser.get('SERVER','%s.pkeyfile'%servname))
        if not os.path.exists(pkey):
            self.write('ERROR:pkeyfile not exists or chmod 600 %s user %s'%(pkey,getpass.getuser()))
            return
        if int(oct(os.stat(pkey).st_mode)[-3:]) != 600:
            self.write('MESSAGE:pls chmod 600 %s user %s'%(pkey,getpass.getuser()))
            return
        bpre=parser.get('SERVER','%s.branchpx'%servname)
        port=int(parser.get('SERVER','%s.port'%servname))



        sdir=parser.get('REPO','%s.syncdirs'%dst)
        item=parser.get('REPO','%s.projitem'%dst)
        cper=parser.get('REPO','%s.permissi'%dst)
        if not re.match(bpre.replace(',','|'),cid):
            self.write('ERROR:branch not right')
            return
        #for b in bpre.split(','):
        #    if not re.match(b,cid):
        #        self.write('ERROR:branch not right')
        #        return

        workdirs=parser.get('PATH','workdirs')
        verconfile=os.path.join(workdirs,'vercnf')
        verconf=CappConfig(verconfile)
        try:
            vers=verconf.get('REPO','%s.versions'%dst)
        except:
            vers=''



        dskey=os.path.join(keyd,parser.get('PATH','deploykf'))
        if not os.path.exists(dskey):
            self.write('ERROR:deploykf not exists or chmod 600 %s user %s'%(dskey,getpass.getuser()))
            return
        if int(oct(os.stat(dskey).st_mode)[-3:]) != 600:
            self.write('MESSAGE:pls chmod 600 %s user %s'%(dskey,getpass.getuser()))
            return
        try:
            authk = GittleAuth(pkey=dskey)
        except:
            self.write("ERROR:deploy key valid")
            return

        try:
                repo = Gittle(os.path.join(workdirs,dst), origin_uri=repo_url[item],auth=authk)
        except:
            try:
                repo=Gittle.clone(repo_url[item],os.path.join(workdirs,dst),auth=authk)
            except:
                self.write('ERROR:deploy key not permission or confdirs cannot write')
                return
        try:
            repo.switch_branch('master')
            repo.pull_from(repo_url[item],cid)
            repo.switch_branch(cid)
            lastcommit=repo.commit_info(0,1,cid)[0]
            self.write('%s\t%s\n%s\n'%(lastcommit['summary'][:60],lastcommit['sha'][:8],lastcommit['committer']['name']))
        except:
            try:
                rmtree(os.path.join(workdirs,dst))
                repo=Gittle.clone(repo_url[item],os.path.join(workdirs,dst),auth=authk)
                repo.switch_branch('master')
                repo.pull_from(repo_url[item],cid)
                repo.switch_branch(cid)
                lastcommit=repo.commit_info(0,1,cid)[0]
                self.write('%s\t%s\n%s\n'%(lastcommit['summary'][:60],lastcommit['sha'][:8],lastcommit['committer']['name']))
            except Exception as e:
                self.write('confdirs cannot write')
                return

        parser2=CappConfig(parser.get('PATH','excludef'))
        try:
            parser2.items('%s%s%s'%(servname,item,dst))
        except:
            self.write('ERROR:excludef no section %s%s%s'%(servname,item,dst) )
            return
        self.write('SETP1 \tstart:\n')
        for sfile,dfile in parser2.items('%s%s%s'%(servname,item,dst)):
            ssdir=os.path.join(parser.get('PATH','confdirs'),'%s%s%s'%(servname,item,dst))
            ddir=os.path.join(workdirs,dst)
            if dfile.startswith('/'):dfile=dfile[1:]
            err,out=rsync(os.path.join(ssdir,sfile),os.path.join(ddir,dfile))

            if len(err) is 0:
                reChar=re.compile('send*.*list|send*.*sec|total size*.*|rsync:')

                self.write('%s'%reChar.sub('',out))

            else:
                reChar=re.compile('rsync error:*.*|rsync: link_stat|rsync:')
                self.write('ERROR:%s'%reChar.sub('',err))
                return
        self.write('SETP1 \tend.\n')

        conflist=os.path.join(ddir,'.git/conflist')
        if os.path.exists(conflist):
            rsync_arg='--exclude=.git --exclude-from=%s'% os.path.join(ddir,'.git/conflist')
        else:
            rsync_arg='--exclude=.git'

        local_dir='%s/'%ddir
        remote_dir='%s@%s:%s'%(user,host,sdir)

        err,_=sshto(host,int(port),user,pwds,pkey).exe_cmdOne('')
        if err:
            self.write('ERROR:target server auth failure')
            return

        only_com=False
        try:
            precmd,aftcmd=parser.get('REPO','%s.cmdexecu'%dst).split(',')
        except:
            precmd=parser.get('REPO','%s.cmdexecu'%dst);aftcmd=''
        if os.path.isfile(precmd):precmd=open(precmd).read()
        if os.path.isfile(aftcmd):aftcmd=open(aftcmd).read()

        if self.get_argument('check',None) is None:
            if not self.get_current_user()['username'] + '|c' in cper:
                if precmd:
                    precmd_err,precmd_out=sshto(host,int(port),user,pwds,pkey).exe_cmdOne(precmd)
                    if len(precmd_err) is 0:
                        self.write('EXEC PRE:\n%s\n'%precmd_out)
                    else:
                        self.write('ERROR: exec pre\n%s\n'%precmd_err)
                        return

                err2,out2= rsync(local_dir,remote_dir,arg=rsync_arg,pkey=pkey)

                if aftcmd:
                    aftcmd_err,aftcmd_out=sshto(host,int(port),user,pwds,pkey).exe_cmdOne(aftcmd)
                    if len(aftcmd_err) is 0:
                        self.write('EXEC AFT:\n%s\n'%aftcmd_out)
                    else:
                        self.write('ERROR: exec aft\n%s\n'%aftcmd_err)


                if out2:
                    config=ConfigParser.RawConfigParser()
                    if vers:
                        vers=vers.split(',')
                        if len(vers) ==1:vers.append(cid)
                        vers[1]=vers[0];vers[0]=cid

                        config.read(verconfile)
                        config.set('REPO','%s.versions'%dst,','.join(vers))
                        with open(verconfile,'w') as configfile:
                            config.write(configfile)
                    else:
                        config.read(verconfile)
                        config.set('REPO','%s.versions'%dst,cid)
                        config.write(open(verconfile,'w'))


            else:
                self.write('ERROR:no publishing rights\n')
                return

        else:
            self.write('EXEC PRE:\n%s\n'%precmd)
            err2,out2=rsync(local_dir,remote_dir,arg=rsync_arg,pkey=pkey,check=True)
            self.write('EXEC AFT:\n%s\n'%aftcmd)
            only_com=True


        if len(err2) is 0:
            reChar=re.compile('send*.*list|send*.*sec|total size*.*|rsync:')
            self.write('SETP2 \tstart:\n')
            self.write('%s'%reChar.sub('',out2))
            self.write('SETP2 \tend.\n')
        else:
            reChar=re.compile('rsync error:*.*|rsync: link_stat|rsync:')
            self.write('%s'%reChar.sub('',err2))

        if only_com:
            self.write('MESSAGE:only for comparison is not pub\n')
        return


    def get(self):
        self.redirect("/pbsys")



def main():
    parse_command_line()
    app = tornado.web.Application(
        [
            (r"/", Publish),
            (r"/pbsys", MainHandler),
            (r"/pbsys/pub", Publish),
            (r"/pbsys/auth/login", AuthLoginHandler),
            (r"/pbsys/auth/logout", AuthLogoutHandler),
            ],
        cookie_secret="6aaee24a70df8615579c355b47d2623f3f7296ee81667f275a91c2cbde2e093faa1dc894ba60cc1bbcfaa40bf910e8a05b7714e21503f48135b61d46b71ddaab",
        login_url="/pbsys/auth/login",

        template_path=os.path.join(os.path.dirname(__file__), "templates"),
        static_path=os.path.join(os.path.dirname(__file__), "static"),
        xsrf_cookies=True,
        debug = False,
        )
    app.listen(options.port,options.address)
    tornado.ioloop.IOLoop.instance().start()


define("port", default=3001, help="run on the given port", type=int)
define("address", default='127.0.0.1', help="run on the given listen address", type=str)
if len(argv)>1:
    if ':' in argv[1]:

        address,port=argv[1].split(':')
        options.address=str(address)
        options.port=int(port)
    else:
        options.address=str(argv[1])




if __name__ == "__main__" :
    try:
        main()
    except Exception as e:
        print e
        exit(1)










