#!/usr/bin/env bash

## set up three docker containers
UUID1=$(cat /proc/sys/kernel/random/uuid)
UUID2=$(cat /proc/sys/kernel/random/uuid)
UUID3=$(cat /proc/sys/kernel/random/uuid)

docker run -d -v ./test:/conf --name $UUID1 nanobox/yoke yoke /conf/primary.conf
docker run -d -v ./test:/conf --name $UUID2 nanobox/yoke yoke /conf/secondary.conf
docker run -d -v ./test:/conf --name $UUID3 nanobox/yoke yoke /conf/monitor.conf


exit



#########################################
# Configuration
#########################################
DB_HOST='192.168.2.100'
DB_USER='postgres'
DB_NAME='datatest'
NUM_USER=10
NUM_BCKT=10
NUM_OBJ=50
SLEEP_DELAY=.1
LONG_SLEEP=.5
PGKILL_SLEEP=45
PGKILL_SIGNAL='SIGINT'
YKILL_SLEEP=120
LOGFILE='/var/tmp/data-test.log'
#########################################
CREATEDB="CREATE DATABASE ${DB_NAME} WITH TEMPLATE = template0 OWNER = ${DB_USER};\n\\\connect ${DB_NAME}\nCREATE TABLE buckets (    id uuid NOT NULL,    name character varying(100) NOT NULL,    user_id uuid NOT NULL);\nCREATE TABLE objects (    id uuid NOT NULL,    alias character varying(255) NOT NULL,    size bigint,    bucket_id uuid NOT NULL);\nCREATE TABLE users (    id uuid NOT NULL,    key character(10) NOT NULL,    admin boolean DEFAULT false,    maxsize bigint DEFAULT 0);\nALTER TABLE ONLY buckets    ADD CONSTRAINT buckets_pkey PRIMARY KEY (id);\nALTER TABLE ONLY buckets    ADD CONSTRAINT buckets_user_id_name_key UNIQUE (user_id, name);\nALTER TABLE ONLY objects    ADD CONSTRAINT objects_bucket_id_alias_key UNIQUE (bucket_id, alias);\nALTER TABLE ONLY objects    ADD CONSTRAINT objects_pkey PRIMARY KEY (id);\nALTER TABLE ONLY users    ADD CONSTRAINT users_pkey PRIMARY KEY (id);\nALTER TABLE ONLY buckets    ADD CONSTRAINT buckets_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);\nALTER TABLE ONLY objects    ADD CONSTRAINT objects_bucket_id_fkey FOREIGN KEY (bucket_id) REFERENCES buckets(id);\n"

# Smart data insertion
pg_query(){
  psql -U ${DB_USER} -d ${DB_NAME} -h ${DB_HOST} -et -c "${1}"

  # If insert failed, try again
  while [[ $? -ne 0 ]]; do
    sleep ${LONG_SLEEP}
    psql -U ${DB_USER} -d ${DB_NAME} -h ${DB_HOST} -et -c "${1}"
  done;
}

# Start chaos monkeys
start_monkeys(){
  # Kill postgres on master.
  while true; do
    sleep ${PGKILL_SLEEP}
    echo "[ `date '+%b %d %H:%M:%S'` ] Attempting to kill postgresql on master" >> {LOGFILE}
    >&2 echo "[ `date '+%b %d %H:%M:%S'` ] Attempting to kill postgresql on master"
    ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i /home/postgres/.ssh/id_rsa postgres@192.168.2.100 "for i in \$(pgrep postgres); do kill -${PGKILL_SIGNAL} \$i; done" >> {LOGFILE} 2>&1
  done &
  pgres_pid=$!
  # Demote master
  while true; do
    sleep ${YKILL_SLEEP}
    echo "[ `date '+%b %d %H:%M:%S'` ] Attempting to demote yoke" >> {LOGFILE}
    >&2 echo "[ `date '+%b %d %H:%M:%S'` ] Attempting to demote yoke"
    yokeadm demote -h ${DB_HOST} -p 4400 >> {LOGFILE} 2>&1
  done &
  yoke_pid=$!
}

# Stop chaos monkeys
kill_monkeys(){
  [[ -n $yoke_pid ]]  && kill $yoke_pid
  [[ -n $pgres_pid ]] && kill $pgres_pid
}

# Start a fresh test
start_test(){
  >&2 echo "[ `date '+%b %d %H:%M:%S'` ] Clearing database"
  pg_query 'TRUNCATE users, buckets, objects CASCADE;' 2>&1

  user=0
  while [ $user -lt $NUM_USER ]; do
    sleep ${SLEEP_DELAY}
    # Add users
    user_id=`uuid -v4`
    user_key="useridis-${user}"
    pg_query "INSERT INTO users VALUES ('${user_id}','${user_key}','f');"  2>&1
    >&2 echo "[ `date '+%b %d %H:%M:%S'` ] Added user '$(( $user +1 ))' of '${NUM_USER}'"
    (( user++ ));
    
    bucket=0
    # Add buckets
    while [ $bucket -lt $NUM_BCKT ]; do
      sleep ${SLEEP_DELAY}
      bucket_id=`uuid -v4`
      bucket_name="bucket${bucket}"
      pg_query "INSERT INTO buckets VALUES ('${bucket_id}','${bucket_name}','${user_id}');" 2>&1
      (( bucket++ ));
    
      object=0
      # Add objects
      while [ $object -lt $NUM_OBJ ]; do
        sleep ${SLEEP_DELAY}
        object_id=`uuid -v4`
        object_alias="object${object}"
        object_size="${object}"
        pg_query "INSERT INTO objects VALUES ('${object_id}','${object_alias}',${object_size},'${bucket_id}');" 2>&1
        (( object++ ));

      done
    done
  done
  >&2 echo "[ `date '+%b %d %H:%M:%S'` ] Test done"
}


# Check results
check_results(){
  num_users=`pg_query 'SELECT COUNT(*) FROM users;' | tr -d '\n| ' | cut -f2 -d';'`
  num_buckets=`pg_query 'SELECT COUNT(*) FROM buckets;' | tr -d '\n| ' | cut -f2 -d';'`
  num_objects=`pg_query 'SELECT COUNT(*) FROM objects;' | tr -d '\n| ' | cut -f2 -d';'`

  if [[ $1 == 'forced' ]] && [[ -n $object ]]; then
    expected_users=$(( $user ))
    expected_buckets=$(( (($user - 1) * $NUM_BCKT) + $bucket ))
    expected_objects=$(( (($expected_buckets - 1) * $NUM_OBJ) + $object ))
  else
    expected_users=$NUM_USER
    expected_buckets=$(( $NUM_USER * $NUM_BCKT ))
    expected_objects=$(( $NUM_USER * $NUM_BCKT * $NUM_OBJ ))
  fi

  if [ $num_users -ne $expected_users ]; then
    echo -en "\e[0;31mDATA LOSS in 'users' table, got '${num_users}', expecting '${NUM_USER}'.\e[0m\n"
  else
    echo -en "\e[0;32mAll data accounted for in 'users'.\e[0m\n"
  fi
  if [ $num_buckets -ne $expected_buckets ]; then
    echo -en "\e[0;31mDATA LOSS in 'buckets' table, got '${num_buckets}', expecting '${NUM_BCKT}'.\e[0m\n"
  else
    echo -en "\e[0;32mAll data accounted for in 'buckets'.\e[0m\n"
  fi
  if [ $num_objects -ne $expected_objects ]; then
    echo -en "\e[0;31mDATA LOSS in 'objects' table, got '${num_objects}', expecting '${NUM_OBJ}'.\e[0m\n"
  else
    echo -en "\e[0;32mAll data accounted for in 'objects'.\e[0m\n"
  fi
}

# Clear logfile
echo > ${LOGFILE}

2>&1 echo -en ${CREATEDB} | psql -U ${DB_USER} -h ${DB_HOST} >> ${LOGFILE}

1>&2
trap '1>&2 echo -en "\nSignal Caught!\n"; kill_monkeys; 1>&2 check_results "forced"; exit 2' SIGINT SIGTERM

start_monkeys >> ${LOGFILE}

time start_test >> ${LOGFILE}

kill_monkeys

check_results | tee -a ${LOGFILE}