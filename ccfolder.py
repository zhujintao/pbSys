import ConfigParser
class CappConfig(ConfigParser.ConfigParser):
    def __init__(self,filename):
        ConfigParser.ConfigParser.__init__(self)
        self.filename=filename
        self.read(filename)
    def optionxform(self,optionstr):
        return optionstr


parser=CappConfig('conf.ini')
#parser.itemso('REPO')
aa=[]
for s in parser.items('REPO'):
    if '.projitem' in s[0]:
       aa.append(s[0].split('.projitem')[0])



import os
for dst  in aa:
   item=parser.get('REPO','%s.projitem'%dst)
   servname=parser.get('REPO','%s.servname'%dst)
   try:
   	os.mkdir('%s/%s%s%s'%(parser.get('PATH','confdirs'),servname,item,dst))
   except:
        pass

    

