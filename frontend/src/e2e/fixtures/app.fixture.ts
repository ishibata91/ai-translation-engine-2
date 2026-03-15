import {test as base} from '@playwright/test';
import {installWailsMocks} from '../helpers/wails-mock';
import {LayoutShellPO} from '../page-objects/components/layout-shell.po';
import {DashboardPO} from '../page-objects/pages/dashboard.po';
import {DictionaryBuilderPO} from '../page-objects/pages/dictionary-builder.po';
import {MasterPersonaPO} from '../page-objects/pages/master-persona.po';
import {TranslationFlowPO} from '../page-objects/pages/translation-flow.po';

type AppPageObjects = {
  dashboard: DashboardPO;
  dictionaryBuilder: DictionaryBuilderPO;
  layoutShell: LayoutShellPO;
  masterPersona: MasterPersonaPO;
  translationFlow: TranslationFlowPO;
};

export const test = base.extend<{ app: AppPageObjects }>({
  app: async ({page}, provideApp) => {
    await installWailsMocks(page);

    await provideApp({
      dashboard: new DashboardPO(page),
      dictionaryBuilder: new DictionaryBuilderPO(page),
      layoutShell: new LayoutShellPO(page),
      masterPersona: new MasterPersonaPO(page),
      translationFlow: new TranslationFlowPO(page),
    });
  },
});
