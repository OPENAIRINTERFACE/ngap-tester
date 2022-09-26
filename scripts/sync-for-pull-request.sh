# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim dockerfile-optimization \ #1367fe7d1b3b59e7ca58b49cd2ab6a416daa8c26
   --nas ngap-tester \
   --ngap ngap-tester
