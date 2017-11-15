export class ServiceStep2Type {
  image_name: string;
  image_tag: string;
  project_id: number;
  project_name: string;
  image_template: string;
}

export class ImageDockerfile {
  image_base: string;
  image_author: string;
  image_volume?: Array<string>;
  image_copy?: Array<{dockerfile_copyfrom?: string, dockerfile_copyto?: string}>;
  image_run?: Array<string>;
  image_env?: Array<{dockerfile_envname?: string, dockerfile_envvalue?: string}>;
  image_expose?: Array<string>;
  image_entrypoint?: string;
  image_cmd?: string;

  constructor() {
    this.image_base = "";
    this.image_volume = Array<string>();
    this.image_run = Array<string>();
    this.image_expose = Array<string>();
    this.image_copy = Array<{dockerfile_copyfrom?: string, dockerfile_copyto?: string}>();
    this.image_env = Array<{dockerfile_envname?: string, dockerfile_envvalue?: string}>();
  }
}

export class ServiceStep2NewImageType extends ServiceStep2Type {
  image_dockerfile: ImageDockerfile;

  constructor() {
    super();
    this.image_dockerfile = new ImageDockerfile();
  }
}

export interface FactoryByPropertyName {
  getInstanceByPropertyName(propName: string): Object
}

/*---------------------------------service configure start-------------------------------------*/
export class ContainerPort {
  hostPort?: number = 0;     //old=>service_nodeport
  containerPort?: number = 0;//old=>container_ports
}

export class EnvVar {
  name: string = ""; //old=>env_name
  value: string = "";//old=>env_value
}

export class VolumeMount {
  name: string = "";          //old=>target_storagename
  readOnly?: boolean = false;
  mountPath: string = "";     //old=>container_dir
  subPath?: string = "";
  ui_nfs_server: string = ""; //old=>target_storageServer;only for ui;
  ui_nfs_path: string = "";   //old=>target_target_dir;only for ui;
}

export class Container implements FactoryByPropertyName {
  name: string = "";               //old=>container_name
  image: string = "";              //old=>container_baseimage
  command: Array<string> = Array();//old=>container_command
  args: Array<string> = Array();
  workingDir: string = "";         //old=>container_workdir
  env: Array<EnvVar> = Array();    //old=>container_envs
  ports: Array<ContainerPort> = Array();
  volumeMounts: Array<VolumeMount> = Array();

  getInstanceByPropertyName(propName: string): Object {
    switch (propName) {
      case "env":
        return new EnvVar();
      case "ports":
        return new ContainerPort();
      case "volumeMounts":
        return new VolumeMount();
      default:
        return {};
    }
  }
}

export type ServiceContainerList = Array<Container>;

export class ProjectInfo {
  service_id: number = 0;
  project_id: number = 0;
  service_name: string = "";
  project_name: string = "";
  namespace: string = "";
  comment: string = "";
  config_phase: string = "";
  service_externalpath: Array<string> = Array();//old=>service_external.service_externalpath
}

export class ObjectMeta {
  name: string = "";
  namespace: string = "";
  labels: {[key: string]: string} = {app: ""};
}

export class HostPathVolumeSource {
  path: string = "";
}

export class EmptyDirVolumeSource {
  medium: string = "";
}

export class NFSVolumeSource {
  server: string = ""; //old=>target_storageServer
  path: string;        //old=>target_dir
  ReadOnly?: boolean = false;
}

export class Volume {
  name: string = "";
  hostPath?: HostPathVolumeSource = new HostPathVolumeSource();
  emptyDir?: EmptyDirVolumeSource = new EmptyDirVolumeSource();
  nfs: NFSVolumeSource = new NFSVolumeSource();
}

export class PodSpec implements FactoryByPropertyName {
  volumes: Array<Volume> = Array();
  containers: Array<Container> = Array();

  getInstanceByPropertyName(propName: string): Object {
    switch (propName) {
      case "volumes":
        return new Volume();
      case "containers":
        return new Container();
      default:
        return {};
    }
  }
}

export class PodTemplateSpec {
  metadata: ObjectMeta = new ObjectMeta();//only input labels
  spec: PodSpec = new PodSpec();
}

export class ReplicationControllerSpec {
  replicas: number = 1;                   //old=>deployment_replicas
  selector: {[key: string]: string} = {};
  template: PodTemplateSpec = new PodTemplateSpec();
}

export class ReplicationController {
  readonly kind: string = "Deployment";              //fixed value
  readonly apiVersion: string = "extensions/v1beta1";//fixed value
  metadata: ObjectMeta = new ObjectMeta();           //only input name value old=>deployment_name || service_name
  spec: ReplicationControllerSpec = new ReplicationControllerSpec();
}

export class ServicePort {
  name: string = "";              //old=>service_external.service_containername
  port: number = 0;               //old=>service_external.service_containerport
  nodePort: number = 0;           //old=>service_external.service_nodeport
}

export class ServiceSpec implements FactoryByPropertyName {
  ports: Array<ServicePort> = Array();   //old=>service_external
  selector: {[key: string]: string} = {app: ""};
  type: string = "";                    //ports.length > 0? =>"NodePort":""

  getInstanceByPropertyName(propName: string): Object {
    switch (propName) {
      case "ports":
        return new ServicePort();
      default:
        return {};
    }
  }
}

export class Service {
  readonly kind: string = "Service";    //fixed value
  readonly apiVersion: string = "v1";   //fixed value
  metadata: ObjectMeta = new ObjectMeta;//metadata.name = service_name; service_name.labels={"app":service_name}
  spec: ServiceSpec = new ServiceSpec();
}

export class DeploymentServiceData {        //equal ServiceConfig2 on goLang
  deployment_yaml: ReplicationController = new ReplicationController();
  service_yaml: Service = new Service();
  projectinfo: ProjectInfo = new ProjectInfo();
}

export class ServiceStep6Output {
  service_id: number;
  service_name: string;
  project_id: number;
  project_name: string;
  service_owner: string;
  service_creationtime: string;
  service_public: number;
}
/*---------------------------------service configure end-------------------------------------*/
