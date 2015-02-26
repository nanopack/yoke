# yoke
postgres redundant auto fail over solution 

for this to run it needs a few things to be in place first:

* primary and secondary needs postgres installed and in the path
* primary and secondary needs to be able to ssh with eachother without passwords
* primary and secondary needs rsync installed or a alternative sync_command 

