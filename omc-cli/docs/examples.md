# Examples
- Retrieving master nodes by label:
```
$ omc get node -l node-role.kubernetes.io/master= -o name   
node/ip-10-0-132-49.eu-central-1.compute.internal
node/ip-10-0-178-163.eu-central-1.compute.internal
node/ip-10-0-202-187.eu-central-1.compute.internal
```
- Retrieving etcd pod name from node name:
```
$ omc get pod -l app=etcd -o jsonpath="{.items[?(@.spec.nodeName=='ip-10-0-132-49.eu-central-1.compute.internal')].metadata.name}"
etcd-ip-10-0-132-49.eu-central-1.compute.internal
```
- Check the ETCD status:
```
$ omc etcd status
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|         ENDPOINT          |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| https://10.0.202.187:2379 | 9f38784f0a8ae43  | 3.4.14  | 147 MB  | false     | false      |        24 |    5682273 |            5682273 |        |
| https://10.0.132.49:2379  | 83b81478d4b02409 | 3.4.14  | 148 MB  | false     | false      |        24 |    5682423 |            5682423 |        |
| https://10.0.178.163:2379 | dd17c7ce8efc0349 | 3.4.14  | 147 MB  | true      | false      |        24 |    5682537 |            5682537 |        |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```
- Retrive the prometheus alerts in `firing` or `pending` state:
```
$ omc prometheus alertrule -s firing,pending -o wide
GROUP                        RULE                                 STATE     AGE   ALERTS   ACTIVE SINCE
cluster-version              UpdateAvailable                      firing    11s   1        27 Jan 22 14:32 UTC
logging_fluentd.alerts       FluentdQueueLengthIncreasing         pending   27s   1        29 Jan 22 11:48 UTC
general.rules                Watchdog                             firing    11s   1        25 Jan 22 08:50 UTC
openshift-kubernetes.rules   AlertmanagerReceiversNotConfigured   firing    5s    1        25 Jan 22 08:51 UTC
```