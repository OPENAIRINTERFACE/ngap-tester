# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: N/A
# nas branch: N/A
# ngap branch: N/A
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim upstream-2022-10-20-sync \
   --nas ngap-tester \
   --ngap ngap-tester
