# Specify the branches or commits to be used by the CI workflows
# for each sub-module
# If you want to re-trigger the CI jobs with the same branches
# But new commits, just add the commit SHA-ONE as comment
# gnbsim branch: ci-adding-golangci-lint # d733a6def588def03422d3809b427cff0bd75dda
# nas branch: test-go-linter # 7cfeb68a8a2d8f3b64705250109a97e568838760
# ngap branch: test-go-linter # eb97583fcd59ab76f2e6ae9adfbef299e399e7a7
./scripts/sync-sub-modules.py \
   --synchronize \
   --force \
   --gnbsim ci-adding-golangci-lint \
   --nas test-go-linter \
   --ngap test-go-linter
