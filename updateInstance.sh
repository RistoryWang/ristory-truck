#description: bash to update instance
prefixPath=/home/deployer/server
serverName=$1
newVersion=$2
instance=$3
downloadPrefix=ftp://192.168.10.168
if [ -z $instance ]; then
  instanceWithHyphen=""
  echo "0.instanceId is null"
else
    echo "0.instanceId is "$3
    if [ $instance -eq 1 -o $instance -eq 2 ]; then
      instanceWithHyphen=_${instance}
    else
      echo "0.incorrect instanceId!"
      exit 0
    fi
fi
instanceName=${serverName}${instanceWithHyphen}
echo "0.instanceName is "$instanceName
newPath=${prefixPath}/${instanceName}/${serverName}-${newVersion}
echo "0.newPath is "$newPath
confPath=/home/deployer/conf/${instanceName}
echo "0.confPath is "$confPath
if [ -d ${confPath} ];then
    echo ${confPath}
else
    echo "0.incorrect confPath:"${confPath}
    exit 0
fi
if [ $# -eq 2 -o $# -eq 3 ];then
    echo "0.params num:"$#
else
    echo "0.incorrect params num:"$#
    exit 0
fi
mkdir -p ${prefixPath}/${instanceName}
cd ${prefixPath}/${instanceName}
if [ -d ${newPath} ];then
    echo "1.don't need download zipfile,exist newPath:"${newPath}
else
    wget ${downloadPrefix}/${serverName}-${newVersion}.zip
    if [ -e ${newPath}.zip ];then
        echo "1.download completed:"${serverName}-${newVersion}.zip
        unzip -q -o ${serverName}-${newVersion}.zip
        echo "1.unzip completed:"${newPath}
    else
        echo "1.incorrect newFile:"${newPath}.zip
        exit 0
    fi
fi
instanceNameFG=${instanceName}
if [[ "$instanceNameFG" != "lss-access" ]];then
  instanceNameFG=${instanceName}  
else
  instanceNameFG="lss-access/"
fi
ptype=$(ps -ef | grep ${instanceNameFG} | grep ${instanceNameFG} | grep -v grep | awk '{print $8}')
if [ -z $ptype ];then
    oldVerIsRun=false
    if [ -e ${confPath}/app.properties ];then
        ptype="java"
    elif [ -e ${confPath}/prod.yml ];then
        ptype="node"    
    else
        echo "2.unknown ptype"
        exit 0
    fi
else
    oldVerIsRun=true
fi
echo "2.ptype is "$ptype
if [[ $(echo $ptype | grep "java") != "" ]];then
    if $oldVerIsRun; then
        pid=$(ps -ef|grep ${instanceNameFG}|grep ${instanceNameFG}|grep -v grep|awk '{print $2}')
        echo "2.pid is "$pid
        pidPath=$(find . -name "*.pid" | xargs grep $pid -ls )
        echo "2.pidPath is "$pidPath
        runPath=$(echo $pidPath | awk -F/ '{print $2}' | cut -d'／' -f2)
        echo "2.runPath is "$runPath
        cd ${runPath}
        ./startServer.sh stop
        echo "2.${instanceName}/${runPath} stopped"
    else
        echo "2.not exist old version running"
    fi
    rm -rf ${newPath}/conf/*.properties
    cp ${confPath}/*.properties ${newPath}/conf
    cd ${newPath}
    ./startServer.sh start
    echo "2.${instanceName}/${serverName}-${newVersion} started"
elif [[ $(echo $ptype | grep "node") != "" ]];then
    rm -rf ${newPath}/app/config/env/*.yml
    cp ${confPath}/*.yml ${newPath}/app/config/env
    cd ${newPath}
    npm rebuild
    export HOSTWEB=$(ifconfig eth0 | grep 'inet'| grep -v '127.0.0.1' | cut -d: -f2 | awk '{ print $2}')
    echo $HOSTWEB
    if $oldVerIsRun; then
        pm2 stop ${instanceName}
        echo "2.${instanceName} stopped"
        pm2 delete ${instanceName}
    else
        echo "2.not exist old version running"
    fi
    NODE_ENV=prod IP=$HOSTWEB pm2 start $serverName --name=${instanceName}
    echo "2.${instanceName}-${newVersion} started"
elif [ -z $ptype ];then
    echo "2.ptype is null"
else
    echo "2.unknown ptype, ptype is"$ptype
fi