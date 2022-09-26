# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: 5fbe198c2249a18566bd7bdd0a5e0fa5ee944fb7
# nas branch: N/A
# ngap branch: N/A
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim dockerfile-optimization \
   --nas ngap-tester \
   --ngap ngap-tester
