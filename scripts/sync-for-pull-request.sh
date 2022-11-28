# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: enable-bypass-traffic-test / 2cfe11cf44980ec15575cbb777585f24ebfde9a4
# nas branch: N/A
# ngap branch: N/A
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim enable-bypass-traffic-test \
   --nas ngap-tester \
   --ngap ngap-tester
