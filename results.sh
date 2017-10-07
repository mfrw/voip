for i in $(ls REC_* | sort -V); do txt=$(./pesq +8000 51.wav $i | tail -1); printf "$i : $txt\n" >> results.txt; echo $i; done
