#!/bin/sh

cd /flyway/sql/ddl

[ -n "$PASSWORD" ] && OPT="-p$(echo "$PASSWORD" | tr -d '\n')"

for F in $(ls *.sql)
do
    echo "Start process $F"
    mysql "$@" "$OPT" < "$F"
    if [ $? -ne 0 ]; then
        echo "Process $F failed"
                return 1
        else
                echo "Process $F successful"
        fi
done
