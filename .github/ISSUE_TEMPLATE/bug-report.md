---
name: Bug report
about: Create a report to help us improve
title: 'Bug: <Brief description of bug>'
labels: bug
assignees: ''

---

**Version of Azure Schema Operator**
<!--- 
The version of the operator pod. 
Assuming your schema operator is deployed in the default namespace, you can get this version from the container image which the controller is running.
Use the following commands:
`kubectl get deployment -n schema-operator-system schema-operator-controller-manager -o wide` and share the image being used by the manager container.   
-->

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
<Fill in the steps>

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Additional context**
Add any other context about the problem here.
