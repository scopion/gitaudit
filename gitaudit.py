import base64
import datetime
import hashlib
import hmac
import json
import logging
import time
import pymysql
import requests
import gitlab

class GitlabAPI(object):
    def __init__(self, *args, **kwargs):
        self.gl = gitlab.Gitlab('http://git.l/', private_token='3VWNVCpX3cf',api_version='4')

    def get_user_projects(self, userid):
        """
        获取用户所拥有的项目
        :param userid:
        :return:
        """
        projects = self.gl.projects.list(owned=True)
        result_list = []
        for project in projects:
            result_list.append(project.http_url_to_repo)
        return result_list

    def getContent(self, projectID):
        """
        通过项目id获取文件内容
        :param projectID:
        :return:
        """
        projects =self.gl.projects.get(projectID)
        f = projects.files.get(file_path='指定项目中的文件路径', ref='master')
        content = f.decode()
        # print(content)
        return content.decode('utf-8')

def is_not_null_and_blank_str(content):
    if content and content.strip():
        return True
    else:
        return False

def post(data,webhook):
    """
    发送消息（内容UTF-8编码）
    :param data: 消息数据（字典）
    :return: 返回发送结果
    """
    #data = json.dumps(data)
    try:
        print (data)
        print (webhook)
        response = requests.post(webhook, data=data)
        print (response.text)
    except requests.exceptions.HTTPError as exc:
        logging.error("消息发送失败， HTTP error: %d, reason: %s" % (exc.response.status_code, exc.response.reason))
        raise
    except requests.exceptions.ConnectionError:
        logging.error("消息发送失败，HTTP connection error!")
        raise
    except requests.exceptions.Timeout:
        logging.error("消息发送失败，Timeout error!")
        raise
    except requests.exceptions.RequestException:
        logging.error("消息发送失败, Request Exception!")
        raise
    else:
        try:
            result = response.json()
        except:
            logging.error("服务器响应异常，状态码：%s，响应内容：%s" % (response.status_code, response.text))
            return {'errcode': 500, 'errmsg': '服务器响应异常'}
        else:
            return result

class FeishuChatbot(object):
    def __init__(self,webhook, secret=None,):
        super(FeishuChatbot, self).__init__()
        self.headers = {'Content-Type': 'application/json; charset=utf-8'}
        self.times = 0
        self.start_time = time.time()
        self.webhook = webhook
        self.secret = secret
        self.timestamp = int(self.start_time)
        self.string_to_sign = '{}\n{}'.format(self.timestamp, secret)
        self.hmac_code = hmac.new(self.string_to_sign.encode("utf-8"), digestmod=hashlib.sha256).digest()
        # 对结果进行base64处理
        self.sign = base64.b64encode(self.hmac_code).decode('utf-8')

    def send_text(self, msg):
        data = {"msg_type": "text","timestamp":self.timestamp,"sign":self.sign}
        if is_not_null_and_blank_str(msg):
            data["content"] = {"text": msg}
        else:
            logging.error("text类型，消息内容不能为空！")
            raise ValueError("text类型，消息内容不能为空！")
        logging.debug('text类型：%s' % data)
        return post(json.dumps(data),self.webhook)

def getmysqlconn():
    config = {
        'host': '120.552',
        'port': 3306,
        'user': 'root',
        'passwd': '1ndC',
        'db': 'gitalter',
        'charset': 'utf8',
        'cursorclass': pymysql.cursors.DictCursor
    }
    conn = pymysql.connect(**config)
    return conn

def updategroups(conn,i):
    cursor = conn.cursor()
    id = i.id
    name = i.name
    updatetime = datetime.datetime.now()
    sql = "insert into gitgroup(id,name,updatetime) values(%s,%s,%s) " \
          "on duplicate key update name=VALUES(name),updatetime=VALUES(updatetime)"
    values = (id,name,updatetime)
    try:
        # 执行sql语句
        n = cursor.execute(sql, values)
        # print n
        # 提交到数据库执行
        conn.commit()
    except Exception as e:
        print(e)
        # Rollback in case there is any error
        conn.rollback()
        print('update fail,rollback')

def updateusers(conn,i,groupid):
    cursor = conn.cursor()
    id = i.id
    name = i.name
    username = i.username
    state = i.state
    email = i.email
    updatetime = datetime.datetime.now()
    sql = "insert into gitusers(id,name,username,state,email,groupid,updatetime) values(%s,%s,%s,%s,%s,%s,%s) " \
          "on duplicate key update name=VALUES(name),username=VALUES(username),state=VALUES(state)," \
          "email=VALUES(email),groupid=VALUES(groupid),updatetime=VALUES(updatetime)"
    values = (id,name,username,state,email,groupid,updatetime)
    try:
        # 执行sql语句
        n = cursor.execute(sql, values)
        # print n
        # 提交到数据库执行
        conn.commit()
    except Exception as e:
        print(e)
        # Rollback in case there is any error
        conn.rollback()
        print('update fail,rollback')

def getusergroups(conn):
    cursor = conn.cursor()
    sql = "select distinct groupid from gitusers"
    groups = []
    try:
        cursor.execute(sql)
        results = cursor.fetchall()
        for i in results:
            id = i.get('groupid')
            groups.append(id)
        return groups
    except:
        # Rollback in case there is any error
        print('select groupids fail,rollback')

def getgroupusers(conn,id):
    cursor = conn.cursor()
    sql = "select userid from groupuser where groupid=%s"
    userids = []
    try:
        cursor.execute(sql,id)
        results = cursor.fetchall()
        for i in results:
            id = i.get('userid')
            userids.append(id)
        return userids
    except:
        # Rollback in case there is any error
        print('select groupids fail,rollback')

def getprojectusers(conn,id):
    cursor = conn.cursor()
    sql = "select userid from projectuser where projectid=%s"
    userids = []
    try:
        cursor.execute(sql,id)
        results = cursor.fetchall()
        for i in results:
            id = i.get('userid')
            userids.append(id)
        return userids
    except:
        # Rollback in case there is any error
        print('select groupids fail,rollback')

def getprojectgroups(conn,id):
    cursor = conn.cursor()
    sql = "select groupid from projectgroup where projectid=%s"
    groupids = []
    try:
        cursor.execute(sql,id)
        results = cursor.fetchall()
        for i in results:
            id = i.get('groupid')
            groupids.append(id)
        return groupids
    except:
        # Rollback in case there is any error
        print('select groupids fail,rollback')

def getgroupgroups(conn,id):
    cursor = conn.cursor()
    sql = "select prigroup from groupgroup where groupid=%s"
    groupids = []
    try:
        cursor.execute(sql,id)
        results = cursor.fetchall()
        for i in results:
            id = i.get('prigroup')
            groupids.append(id)
        return groupids
    except:
        # Rollback in case there is any error
        print('select groupids fail,rollback')



if __name__ == "__main__":
    webhook = 'https://open.feishu.cn/open-apis/bot/v2/hook/ef0cdc32ceb9'
    secret = '20GbTDsHg'
    xiaoding = FeishuChatbot(webhook, secret=secret)
    at = "<at user_id=\"ou_d1b7074b325\">名字</at>"

    url = 'http://giai'
    token = 'UTPFGhB'
    gl = gitlab.Gitlab(url, token)

    conn=getmysqlconn()
    #获取用户组ids
    usergroups = getusergroups(conn)
    ###更新部门用户
    for g in gl.groups.get(118).subgroups.list(all=True):
        updategroups(conn,g)
        for u in gl.groups.get(g.id).members.list(all=True):
            user = gl.users.get(u.id)
            updateusers(conn,user,g.id)

    #获取所有admin用户
    admins = []
    for ad in gl.groups.get(257).members_all.list(all=True):
        admins.append(ad.id)
###检查异常admin用户
    for u in gl.users.list(all=True):
        isadmin = getattr(u, 'is_admin')
        if isadmin:
            if u.id in admins:
                pass
            else:
                adminmsg = "发现异常admin权限用户" +u.name
                xiaoding.send_text(adminmsg)
    #获取所有运维用户
    for op in gl.groups.get(208).members_all.list(all=True):
        admins.append(op.id)
    #
    ####按用户分组问题权限项目
    #{username:[giturl1,giturl]}

    for group in gl.groups.list(all=True):
        prgroups = [257,208]
        prusers = []
        #prusers.extend(admins)

        ###排除用户组group
        if group.id in usergroups:
            pass
        else:
            ###获取group详情
            group = gl.groups.get(group.id)
            print(group)
            ###从数据库中获取已授权group
            ggs = getgroupgroups(conn,group.id)
            print("group授权group"+str(ggs))
            #加入授权组
            prgroups.extend(ggs)
            #父group,只处理了一层父目录，多层父目录的没处理
            parent_id = getattr(group, 'parent_id')
            if parent_id:
                #父的授权group
                prgroups.extend(getgroupgroups(conn, parent_id))
                #父的授权用户
                prusers.extend(getgroupusers(conn, parent_id))

            #所有授权组的用户
            for prg in prgroups:
                for m in gl.groups.get(prg).members.list(all=True):
                    prusers.append(m.id)
            ###获取gitlab中group，校对group-group
            gpgroups = getattr(group, 'shared_with_groups')
            print(gpgroups)
            for pg in gpgroups:
                if pg.get("group_id") in prgroups:
                    pass
                else:
                    msg = "项目组: " + group.full_path + " 发现异常权限组："+str(pg.get("group_name"))
                    print(msg)
                    xiaoding.send_text(msg)
                #######

            #获取group-user权限表
            ##gusers = getgroupusers(conn,g.id)
            #所有权限人员
            prusers.extend(getgroupusers(conn,group.id))
            prusers.extend(admins)
            prusers = list(set(prusers))
            print(prusers)
            ###权限表只纪录每个group的直属members，对比也只对比每个group的直属members。
            ####校对group-users。
            gmembers = group.members.list(all=True)
            msgids = []
            for gme in gmembers:
                if gme.id in prusers:
                    pass
                else:
                    msgids.append(gme.name)
            if msgids:
                msg = "项目组: " + group.full_path + " 发现异常权限用户："+str(msgids)
                print(msg)
                xiaoding.send_text(msg)


            for p in group.projects.list(with_shared=False):
                project = gl.projects.get(p.id)
                print(project)
                #获取project-group权限表
                pgs = getprojectgroups(conn,p.id)
                #将人加入权限表
                for g in pgs:
                    for u in gl.groups.get(g).members.list(all=True):
                        prusers.append(u.id)
                ###加入授权group
                pgs.extend(prgroups)
                print("project授权group"+str(pgs))
                #获得当前project的group，校对project-group
                pgroups = getattr(project,'shared_with_groups')
                print(pgroups)
                for pg in pgroups:
                    if pg.get("group_id") in pgs:
                        pass
                    else:
                        msg = "项目: " + project.web_url + " 发现异常权限组：" + str(pg.get("group_name"))
                        print(msg)
                        xiaoding.send_text(msg)

                #获取所有组内所有用户，组成一个大表，报警时按用户报出所有project
                #######

                #获取project-user权限表
                dbusers = getprojectusers(conn,p.id)
                prusers = list(set(prusers))
                print("继承授权用户"+str(prusers))
                print("单独授权用户"+str(dbusers))
                ####只获取members，不获取组继承的用户,校对project-user
                pmembers = project.members.list(all=True)
                pmsgids = []
                expids = []
                for pme in pmembers:
                    print(pme)
                    if pme.id in prusers:
                        pass
                    else:
                        exp = getattr(pme, 'expires_at')
                        if exp:
                            pass
                        else:
                            expids.append(pme.name)
                        if pme.id in dbusers:
                            pass
                        else:
                            pmsgids.append(pme.name)
                if pmsgids:
                    msg = "项目: " + project.web_url + " 发现异常权限用户：" + str(pmsgids)
                    xiaoding.send_text(msg)
                if expids:
                    expmsg = "项目: " + project.web_url + " 中用户: " + str(expids) + "未设置过期时间"
                    xiaoding.send_text(expmsg)








