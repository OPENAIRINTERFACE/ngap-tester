# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: upstream-2022-10-20-sync / fb864d5141f0ceab2513c2071c92de2acf8f14c4
# nas branch: N/A
# ngap branch: N/A
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim upstream-2022-10-20-sync \
   --nas ngap-tester \
   --ngap ngap-tester
