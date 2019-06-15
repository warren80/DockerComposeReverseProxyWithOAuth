# DockerCompose ReverseProxy With oAuth and Vouch

This is is the basis for my cluster of websites.  Hopefully I will be adding more modules in the future.

Docker compose spins up the seperate containers.

I am using nginx as a reverse proxy and using oauth to authenticate requests to www.wvoelkl.com.  
The same patern in ./nginx/conf.d/default.conf

will need to be repeated for all new servers.

TODO:
Seperate the Vouch config out.
Configure certbot better
Setup a jenkins build server.

add more applications :)

With Vouch

1. under ./web/site mount a static/dynamic website
  mount --bind src dst or in /etc/fstab src dst none defaults,bind 0 0 
2. mv and modifiy ./config/config.yml-sample to a working version. to ./config/config.yml
3. docker-compose up
