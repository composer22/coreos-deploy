### [Dockerized] (http://www.docker.com) [coresos-deploy](https://registry.hub.docker.com/u/composer22/coresos-deploy/)

A docker image for coresos-deploy. This is created as a single "static" executable using a lightweight image.

To make:

cd docker
./build.sh

Once it completes, you can run the server. Example launch:
```
docker run -v /var/run/fleet.sock:/var/run/fleet.sock --name tester -p 8081:8080 -p 6061:6060 \
 -d composer22/coreos-deploy -N "San Francisco" -H 0.0.0.0 -O example.com -E development -p 8080 \
-X 2 -T 172.21.5.172:2379 \
--dsn "coreos_deploy:mypw3@tcp(your-sql-db.hsj237a8n3ds.us-east-1.rds.amazonaws.com:3306)/coreos_deploy_dev"
```
To get the IP address for the -T option of etcd2, use one the following commands on the running CoreOS instance:
```
ip route show | grep default | awk '{print $9}'
ip route show | grep docker0 | awk '{print $9}'
```
