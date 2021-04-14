FROM node:15.14.0-alpine3.10

RUN mkdir -p /user/node/app/sqedule
WORKDIR /user/node/app/sqedule
COPY . .
RUN npm install
# This should be fixed somehow
#USER node 
EXPOSE 3000

CMD "npm" "run" "dev"