kill `ps -ef | grep "go-build" | grep -v grep | awk '{print $2}'`
cd /data/www/pokiwarh5/src/GoLang
> data/logs/go_run_log.txt
nohup go run /data/www/pokiwarh5/src/GoLang/main.go >> data/logs/go_run_log.txt 2>&1 &
echo 'Restart GoLang API Done!'
rm -rf /data/www/pokiwarh5/src/GoLang/data/sessions/a/*

