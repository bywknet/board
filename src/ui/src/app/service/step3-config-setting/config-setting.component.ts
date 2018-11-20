import { ChangeDetectorRef, Component, Injector, OnInit } from '@angular/core';
import { ValidationErrors } from "@angular/forms/forms";
import { HttpErrorResponse } from "@angular/common/http";
import { Observable } from "rxjs/Observable";
import {
  Container,
  ExternalService,
  PHASE_CONFIG_CONTAINERS,
  PHASE_EXTERNAL_SERVICE,
  ServiceStepPhase,
  UIServiceStep2,
  UIServiceStep3,
  UIServiceStepBase
} from '../service-step.component';
import { ServiceStepBase } from "../service-step";
import { IDropdownTag } from "../../shared/shared.types";
import { SetAffinityComponent } from "./set-affinity/set-affinity.component";
import 'rxjs/add/observable/forkJoin'

@Component({
  styleUrls: ["./config-setting.component.css"],
  templateUrl: './config-setting.component.html'
})
export class ConfigSettingComponent extends ServiceStepBase implements OnInit {
  patternServiceName: RegExp = /^[a-z]([-a-z0-9]*[a-z0-9])+$/;
  patternIP: RegExp = /^((?:(?:25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))\.){3}(?:25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d))))$/;
  showAdvanced = false;
  showNodeSelector = false;
  isActionWip: boolean = false;
  nodeSelectorList: Array<{name: string, value: string, tag: IDropdownTag}>;
  uiPreData: UIServiceStep2;

  constructor(protected injector: Injector,
              private changeDetectorRef: ChangeDetectorRef) {
    super(injector);
    this.changeDetectorRef.detach();
    this.nodeSelectorList = Array<{name: string, value: string, tag: IDropdownTag}>();
    this.uiPreData = new UIServiceStep2();
  }

  ngOnInit() {
    let obsStepConfig = this.k8sService.getServiceConfig(this.stepPhase);
    let obsPreStepConfig = this.k8sService.getServiceConfig(PHASE_CONFIG_CONTAINERS);
    Observable.forkJoin(obsStepConfig, obsPreStepConfig).subscribe((res: [UIServiceStepBase, UIServiceStepBase]) => {
      this.uiBaseData = res[0];
      this.uiPreData = res[1] as UIServiceStep2;
      if (this.uiData.externalServiceList.length === 0 && this.uiPreData.containerHavePortList.length > 0) {
        let container = this.uiPreData.containerHavePortList[0];
        this.addNewExternalService();
        this.setExternalInfo(container, 0);
      }
      this.changeDetectorRef.reattach();
    });
    this.nodeSelectorList.push({name: 'SERVICE.STEP_3_NODE_DEFAULT', value: '', tag: null});
    this.k8sService.getNodeSelectors().subscribe((res: Array<{name: string, status: number}>) => {
      res.forEach((value: {name: string, status: number}) => {
        this.nodeSelectorList.push({
          name: value.name, value: value.name, tag: {
            type: value.status == 1 ? 'alert-success' : 'alert-warning',
            description: value.status == 1 ? 'SERVICE.STEP_3_NODE_STATUS_SCHEDULABLE' : 'SERVICE.STEP_3_NODE_STATUS_UNSCHEDULABLE'
          }
        })
      });
    });
  }

  get stepPhase(): ServiceStepPhase {
    return PHASE_EXTERNAL_SERVICE
  }

  get uiData(): UIServiceStep3 {
    return this.uiBaseData as UIServiceStep3;
  }

  get checkServiceNameFun() {
    return this.checkServiceName.bind(this);
  }

  get nodeSelectorDropdownText() {
    return this.uiData.nodeSelector === '' ? 'SERVICE.STEP_3_NODE_DEFAULT' : this.uiData.nodeSelector;
  }

  get curNodeSelector() {
    return this.nodeSelectorList.find(value => value.name === this.uiData.nodeSelector);
  }

  getContainerDropdownText(index: number): string {
    let result = this.uiData.externalServiceList[index].container_name;
    return result == "" ? "SERVICE.STEP_3_SELECT_CONTAINER" : result;
  }

  setExternalInfo(container: Container, index: number) {
    this.uiData.externalServiceList[index].container_name = container.name;
    this.uiData.externalServiceList[index].node_config.target_port = container.container_port[0];
  }

  setNodePort(index: number, port: number) {
    this.uiData.externalServiceList[index].node_config.node_port = Number(port).valueOf();
  }

  addNewExternalService() {
    if (this.uiPreData.containerHavePortList.length > 0 && !this.isActionWip) {
      let externalService = new ExternalService();
      this.uiData.externalServiceList.push(externalService);
    }
  }

  removeExternalService(index: number) {
    this.uiData.externalServiceList.splice(index, 1);
  }

  setAffinity() {
    if (!this.isActionWip) {
      let factory = this.factoryResolver.resolveComponentFactory(SetAffinityComponent);
      let componentRef = this.selfView.createComponent(factory);
      componentRef.instance.openSetModal(this.uiData).subscribe(() => this.selfView.remove(this.selfView.indexOf(componentRef.hostView)));
    }
  }

  setNodeSelector() {
    if (!this.isActionWip) {
      this.showNodeSelector = !this.showNodeSelector;
    }
  }

  checkServiceName(control: HTMLInputElement): Observable<ValidationErrors | null> {
    return this.k8sService.checkServiceExist(this.uiData.projectName, control.value)
      .map(() => null)
      .catch((err:HttpErrorResponse) => {
        if (err.status == 409) {
          this.messageService.cleanNotification();
          return Observable.of({serviceExist: "SERVICE.STEP_3_SERVICE_NAME_EXIST"});
        } else if (err.status == 404) {
          this.messageService.cleanNotification();
        }
        return Observable.of(null);
      });
  }

  isValidMinNodePort(): boolean {
    if (this.uiData.clusterIp === '') {
      return this.uiData.externalServiceList.every(value => value.node_config.node_port >= 30000);
    } else {
      return this.uiData.externalServiceList.every(value => value.node_config.node_port >= 30000 || value.node_config.node_port == 0)
    }
  }

  haveRepeatNodePort(): boolean {
    let haveRepeat = false;
    this.uiData.externalServiceList.forEach((value, index) => {
      if (this.uiData.externalServiceList.find((value1, index1) =>
        value1.container_name === value.container_name
        && value1.node_config.target_port === value.node_config.target_port
        && index1 !== index)) {
        haveRepeat = true
      }
    });
    return haveRepeat;
  }

  forward(): void {
    if (this.verifyInputValid()) {
      if (this.uiData.externalServiceList.length == 0) {
        this.messageService.showAlert(`SERVICE.STEP_3_EXTERNAL_MESSAGE`, {alertType: "alert-warning"});
      } else if (this.haveRepeatNodePort()) {
        this.messageService.showAlert(`SERVICE.STEP_3_EXTERNAL_REPEAT`, {alertType: "alert-warning"});
      } else if (!this.isValidMinNodePort()){
        this.messageService.showAlert(`SERVICE.STEP_3_INVALID_MIN_NODE_PORT`, {alertType: "alert-warning"});
      } else if (this.uiData.affinityList.find(value => value.services.length == 0)) {
        this.messageService.showAlert(`SERVICE.STEP_3_AFFINITY_MESSAGE`, {alertType: "alert-warning"});
      } else {
        this.isActionWip = true;
        this.k8sService.setServiceConfig(this.uiData.uiToServer()).subscribe(
          () => this.k8sService.stepSource.next({index: 5, isBack: false})
        );
      }
    }
  }

  backUpStep(): void {
    this.k8sService.stepSource.next({index: 2, isBack: true});
  }
}