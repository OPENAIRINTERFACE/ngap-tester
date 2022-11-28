# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: nov-2022-upstream-sync / 1b96757165bb029e942ae75964486a01a523f38a
# nas branch: N/A
# ngap branch: N/A
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim nov-2022-upstream-sync \
   --nas ngap-tester \
   --ngap ngap-tester
