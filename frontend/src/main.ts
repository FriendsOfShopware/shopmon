import { createApp } from 'vue';
import { createPinia } from 'pinia';

import './style.css';

import App from './App.vue';
import { router } from './router';


import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';

/* import FontAwesomeIcon */
import { faKey, faUserPlus, faEnvelope, faCircleInfo, faCircleCheck, faCircleXmark, faXmark, faGear, faPowerOff, faBarsStaggered, faPlus, faFloppyDisk, faTriangleExclamation, faTrash, faCircle, faUpRightFromSquare, faPlug, faListCheck } from '@fortawesome/free-solid-svg-icons';
import { faBell } from '@fortawesome/free-regular-svg-icons';

library.add( faKey, faUserPlus, faEnvelope, faCircleInfo, faCircleCheck, faCircleXmark, faXmark, faBell, faGear, faPowerOff, faBarsStaggered, faPlus, faFloppyDisk, faTriangleExclamation, faTrash, faCircle, faUpRightFromSquare, faPlug, faListCheck );

const app = createApp(App);

app.use(createPinia());
app.use(router);

app.component('font-awesome-icon', FontAwesomeIcon);

app.mount('#app');
