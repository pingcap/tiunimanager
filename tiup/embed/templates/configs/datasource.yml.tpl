apiVersion: 1
deleteDatasources:
  - name: {{.ClusterName}}
datasources:
  - name: {{.ClusterName}}
    id: 1
    uid: 1
    type: prometheus
    access: proxy
    url: http://{{.IP}}:{{.Port}}
    withCredentials: false
    isDefault: false
    tlsAuth: false
    tlsAuthWithCACert: false
    version: 1
    editable: true