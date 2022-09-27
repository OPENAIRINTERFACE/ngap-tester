# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: test-go-linter # b8f6a47ddbb5ea9eb89018edb02a9a1b50221860
# nas branch: test-go-linter # 7cfeb68a8a2d8f3b64705250109a97e568838760
# ngap branch: test-go-linter # eb97583fcd59ab76f2e6ae9adfbef299e399e7a7
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim test-go-linter \
   --nas test-go-linter \
   --ngap test-go-linter
