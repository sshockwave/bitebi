source ~/.bashrc
cd ~/data/bitebi
conda activate go
go run ./... < testcase/peer1.txt
