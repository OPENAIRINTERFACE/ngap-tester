# Specify the branches or commits to be used by the CI workflows
# for each sub-module
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim ngap-tester \
   --nas ngap-tester \
   --ngap ngap-tester
