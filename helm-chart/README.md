### Now you can use this helm-char to run a cronjob with possibility feed prometheus with an pushgateway, serve http/s page of popeye's report.

```
git clone git@github.com:derailed/popeye.git
cd popeye/helm-chart
helm install popeye -n popeye --create-namespace .
``` 

You can set a many values on values.yaml or pass using --set parameter:value, like a helm.

If you have any questios or sugestions please can contact me adonai@ascsystem.com.br
