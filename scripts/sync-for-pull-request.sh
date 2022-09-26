# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: 1367fe7d1b3b59e7ca58b49cd2ab6a416daa8c26
# nas branch: N/A
# ngap branch: N/A
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim dockerfile-optimization \
   --nas ngap-tester \
   --ngap ngap-tester
