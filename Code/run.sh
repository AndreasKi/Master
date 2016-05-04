DEFAULT_PORT=8590
START=1
END=$(($1 - 1))
DAEMONS=$(($END + 1))
COORDMODE=$2

echo Starting $DAEMONS daemons with $COORDMODE coordinator...

cd daemon
./daemon $COORDMODE $DEFAULT_PORT coord &

for (( i=$START; i<=$END; i++ )) do
	sleep .1
	PORT=$(($DEFAULT_PORT + $i))
	./daemon $COORDMODE $PORT &
done