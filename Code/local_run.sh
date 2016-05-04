DAEMONS=$1
COORDMODE=$2
echo Starting $DAEMONS daemons with $COORDMODE coordinator...
cd daemon
for (( i=1; i<=$DAEMONS; i++ )) do
	./daemon $COORDMODE &
	sleep .5
done