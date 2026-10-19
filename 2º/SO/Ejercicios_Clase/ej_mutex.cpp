/*Un túnel de un único sentido tiene limitado a 8 el número de vehículos que pueden estar
simultáneamente dentro. Se pide simular este sistema con hilos, diseñando el algoritmo que
deben ejecutar los hilos para comportarse como los vehículos del ejemplo.*/
mutex_t m;
cond_t c;
int n = 0; int nEsp = 0;

vehiculo()
{
    lock(&m)
    nEsp++;
    while(n == 8){
        cond_wait(&c, &m);
    }
    nEsp--;
    n++;
    unlock(&m);
    Tunel();
    lock(&m);
    n--
    if(nEsp >0){
        cond_signal(&c);
    }
    unlock(&m)
}

/*Se desea simular mediante hilos una situación de sincronización en la que cada hilo representa
a un excursionista. Los excursionistas deben cruzar un puente que solamente soporta 1000
kilos. Diseñar el algoritmo que deben ejecutar los hilos para que, sin que se hunda el puente,
puedan cruzarlo simultáneamente tantos hilos como sea posible.
NOTA: Suponer que cada hilo dispone de una variable local que le indica el peso del
excursionista al que representa.*/
mutex_t m;
cond_t c;
int nEsp = 0; int sum_peso = 0;

puente(int peso)
{
    lock(&m);
    nEsp++;
    while(sum_peso + peso >= 1000)
    {
        cond_wait(&c, &m);
    }
    nEsp--;
    sum_peso = sum_peso + peso;
    unlock(&m);
    puente();
    lock(&m);
    sum_peso = sum_peso - peso;
    if(nEsp > 0){
        cond_broadcast(&c);
    }
    unlock(&m)
}

/*Un túnel de doble sentido tiene limitado a 8 el número de vehículos que pueden estar
simultáneamente dentro. Al túnel llegan vehículos tanto desde el Este como desde el Oeste.
Cuando haya vehículos esperando en ambos lados y quede un hueco libre, debe pasar un
vehículo del Este (es decir, el Este tiene preferencia). Se pide simular este sistema con hilos,
diseñando el algoritmo que deben ejecutar los hilos para comportarse como los coches del
ejemplo.*/
mutex_t m;
cond_t cEste, cOeste;
int n = 0, int nEspE = 0, int nEspO = 0;

este()
{
    lock(&m)
    nEspE++;
    while(n >= 8)
    {
        cond_wait(&cEste, &m);
    }
    nEspE--;
    n++;
    unlock(&m);
    tunel();
    lock(&m);
    n--;
    if(nEspE > 0){
        cond_signal(&cEste);
    } else if(nEspO > 0) {
        cond_signal(&cOeste);
    }
    unlock(&m);
}

oeste()
{
    lock(&m);
    nEspO++;
    while(n>= 8 || nEspE > 0)
    {
        cond_wait(&cOeste , &m);
    }
    nEspO--;
    n++;
    unlock(&m);
    tunel();
    lock(&m);
    n--;
    if(nEspE > 0 ){
        cond_signal(&cEste);
    } else if(nEspO > 0){
        cond_signal(&cOeste);
    }
    unlock(&m);
}

/*El núcleo del sistema operativo Unix utiliza internamente dos primitivas de sincronización
similares a las dos siguientes:
• dormir(valor): bloquear en valor al proceso que la invoca.
• despertar(valor): desbloquear a todos los procesos bloqueados en valor.
Suponiendo que valor es un entero entre 1 y n, proponer una implementación de estas
primitivas utilizando mútex y variables condición como herramienta de sincronización.
Sugerencia: resolver primero el problema como si sólo existiese un valor posible y después
extenderlo al caso de N valores.*/
mutex_t m;
cond_t c;
int n = 0;

dormir()
{
    int miN;
    lock(&m);
    miN = n;
    while(miN == n){
        cond_wait(&c, &m);
    }
    unlock(&m);
}

despertar()
{
    lock(&m);
    n++;
    cond_broadcast(&c);
    unlock(&m);
}

/*En un aeropuerto existe un sistema que permite emitir tarjetas de embarque. Este sistema
tiene una única impresora. En el sistema se ejecutan dos tipos de hilos, los que representan a
los clientes VIP y los que representan a clientes normales. Los hilos VIP tienen preferencia.
Sin embargo, cada 5 clientes VIP se podrá imprimir una tarjeta de un cliente normal.
Usando mutex y variables condición como herramienta de sincronización, proponer un
algoritmo para los hilos VIP y otro para los normales.*/
mutex_t m;
cond_t c, cVIP;
int n = 0; int nEspN = 0, int nEspVIP = 0;

vip()
{
    lock(&m);
    nEspVIP++;
    while(n >= 5){
        cond_wait(&cVIP, &m);
    }
    nEspVIP--;
    n++
    unlock(&m);
    imprimir();
    lock(&m);
    if(nEspVIP > 0 && n < 5){
        cond_signal(&cVIP);
    } else if(n >= 5){
        cond_signal(&c);
    }
    unlock(&m);
}

normal()
{
    lock(&m);
    nEspN++;
    while(n <= 5){
        cond_wait(&c, &m);
    }
    nEspN--;
    n = 0;
    unlock(&m);
    imprimir();
    lock(&m);
    if(nEspVIP > 0){
        cond_signal(&cVIP);
    } else if(nEspN > 0 && n < 5){
        cond_signal(&c);
    }
    unlock(&m);
}

/*En una universidad, hay un laboratorio de computación avanzada que tiene las siguientes reglas de acceso y sincronización:
Capacidad Limitada: En el laboratorio solo pueden entrar un máximo de 5 personas simultáneamente para no sobrecargar el sistema de refrigeración.
Dos tipos de usuarios: Existen hilos que representan a Investigadores (VIP) y otros que representan a Estudiantes (Normales).
Prioridad de Acceso: Los Investigadores tienen preferencia absoluta para entrar al laboratorio. Si hay investigadores esperando, ningún estudiante puede entrar a menos que se cumpla la regla de justicia.
Regla de Justicia (Antistarvation): Para evitar que los estudiantes se queden fuera todo el día, por cada 3 Investigadores que accedan al laboratorio, se debe permitir la entrada de 1 Estudiante (siempre que haya estudiantes esperando y el aforo lo permita).*/
mutex_t m;
cond_t cInvestigador, cEstudiante;
int nEspInv = 0; int nEspEst = 0;
int n = 0; int cont = 0;

investigador()
{
    lock(&m);
    nEspInv++;
    while(n >= 5 || (cont >=3 && nEspEst > 0)){
        cond_wait(&cInvestigador, &m);
    }
    nEspInv--;
    n++;
    cont++;
    unlock(&m);
    laboratorio();
    lock(&m);
    n--;
    if(nEspInv > 0 && cont < 3){
        cond_signal(&cInvestigador);
    } else if(nEspEst > 0 || cont >= 3){
        cond_signal(&cEstudiante);
    } else if(nEspInv > 0){
        cond_signal(&cInvestigador);
    }
    unlock(&m);
}

estudiante()
{
    lock(&m);
    nEspEst++
    while(n >= 5 || (nEspInv > 0 && cont < 3)){
        cond_wait(&cEstudiante, &m);
    }
    nEspEst--;
    n++;
    cont = 0; 
    unlock(&m);
    laboratorio();
    lock(&m);
    n--;
    if(nEspInv > 0 && cont < 3){
        cond_signal(&cInvestigador);
    } else if(nEspEst > 0){
        cond_signal(&cEstudiante);
    } else{
        cond_signal(&cInvestigador);
    }
    unlock(&m);
}

/*Un antiguo castillo medieval tiene un puente levadizo de madera muy estrecho y delicado. Para evitar que se derrumbe, las reglas del guardia son estrictas:
Capacidad: El puente solo soporta un máximo de 3 personas cruzando al mismo tiempo.
Clases Sociales: Hay dos tipos de personas que quieren cruzar: los Nobles (prioridad alta) y los Campesinos (prioridad baja).
Prioridad: Los Nobles tienen preferencia absoluta. Si hay algún Noble esperando para cruzar, los Campesinos deben apartarse y esperar.
Regla del Rey (Justicia): Para evitar revueltas, el Rey ha decretado que, por cada 4 Nobles que crucen el puente de forma consecutiva, el guardia debe dejar pasar obligatoriamente a 1 Campesino (siempre y cuando haya algún campesino esperando en la cola).*/
mutex_t m;
cond_t cNoble, cCampesino;
int nEspN = 0; int nEspC = 0; int contador = 0; int n = 0;

noble()
{
    lock(&m);
    nEspN++;
    while(n >= 3 || contador >=4){
        cond_wait(&cNoble, &m);
    }
    nEspN--;
    n++;
    contador++;
    unlock(&m);
    cruzar_puente();
    lock(&m);
    n--;
    if(nEspN > 0 && contador < 4){
        cond_signal(&cNoble);
    } else if(nEspC > 0 || contador >=4){
        cond_signal(&cCampesino);
    } else if(nEspN >0){
        cond_signal(&cNoble);
    }
    unlock(&m);
}

campesino()
{
    lock(&m);
    nEspC++;
    while(n >= 3 || (nEspN > 0 && contador < 4)){
        cond_wait(&cCampesino, &m);
    }
    nEspC--;
    n++;
    contador = 0;
    unlock(&m);
    cruzar_puente();
    lock(&m);
    n--;
    if(nEspN > 0 && contador < 4){
        cond_signal(&cNoble);
    } else if(nEspC > 0){
        cond_signal(&cCampesino);
    } else if(nEspN > 0) {
        cond_signal(&cNoble);
    }
    unlock(&m);
}

/*Durante una fuerte ventisca, un grupo de personas llega a un pequeño refugio de alta montaña. La sala donde está la chimenea es muy pequeña y el guarda ha establecido unas normas estrictas para evitar agobios:
Aforo: En la sala de la chimenea solo pueden estar un máximo de 4 personas calentándose al mismo tiempo.
Roles: Hay dos tipos de personas: Rescatistas (prioridad alta, necesitan calentarse rápido para volver a salir) y Excursionistas (prioridad baja).
Prioridad: Los Rescatistas tienen preferencia absoluta. Si hay algún Rescatista esperando para entrar, los Excursionistas deben ceder el paso.
Regla de Supervivencia (Justicia): Para evitar que los excursionistas se congelen esperando, el guarda ha decidido que, por cada 3 Rescatistas que entren a la sala de forma consecutiva, se debe dejar pasar obligatoriamente a 1 Excursionista (siempre y cuando haya excursionistas haciendo cola, claro).*/
mutex_t m;
cond_t, cRescatista, cExcursionista;
int n = 0; int nEspR = 0; int nEspE = 0, int contador = 0;

rescatista()
{
    lock(&m);
    nEspR++;
    while(n >= 4){
        cond_wait(&cRescatista, &m);
    }
    nEspE--;
    n++;
    contador++;
    unlock(&m);
    calentarse();
    lock(&m);
    n--;
    if(nEspR > 0 && contador < 3){
        cond_signal(&cRescatista);
    } else if(nEspE > 0 || contador >= 3){
        cond_signal(&cExcursionista);
    } else if(nEspR > 0){
        cond_signal(&cRescatista);
    }
    unlock(&m);
}

excursionista()
{
    lock(&m);
    nEspE++;
    while(n >= 3 || (nEspR > 0 && contador < 3)){
        cond_wait(&cExcursionista, &m);
    }
    nEspE--;
    n++;
    contador = 0;
    unlock(&m);
    calentarse();
    lock(&m);
    n--;
    if(nEspR > 0 && contador < 3){
        cond_signal(&cRescatista);
    } else if(nEspE > 0 || contador >= 3){
        cond_signal(&cExcursionista);
    } else if(nEspR > 0){
        cond_signal(&cRescatista);
    }
    unlock(&m);
}


/*Ejercicio parcial */
mutex_t m;
cond_t cX, cY;
int n = 0, int nEspX = 0,int nEspY = 0; contX = 0; int contY = 0;

X(){
    lock(&m);
    nEspX++;
    while(n>=1 || (nEspY > 0 && contX > 0)){
        cond_wait(&cX, &m);
    }
    nEspX--;
    n++;
    unlock(&m);
    seccion_critica();
    lock(&m);
    n--;
    contX++;
    contY = 0;
    if(nEspY <= 0 || (nEspX > 0 && contY > 0)){
        cond_broadcast(&cX);
    } else if(nEspX <= 0 || (nEspxY > 0 && contX > 0)){
        cond_broadcast(&cY);
    }
    unlock(&m)
}

Y(){
    lock(&m);
    nEspY++;
    while(n>=1 || (nEspX > 0 && contX > 0)){
        cond_wait(&cX, &m);
    }
    nEspY--;
    n++;
    unlock(&m);
    seccion_critica();
    lock(&m);
    n--;
    contY++;
    contX = 0;
    if(nEspY <= 0 || (nEspX > 0 && contY > 0)){
        cond_broadcast(&cX);
    } else if(nEspX <= 0 || (nEspxY > 0 && contX > 0)){
        cond_broadcast(&cY);
    }
    unlock(&m)
}

mutex_t m;
cond_t cNiño, cAdulto;
int nEsp = 0; int aEsp = 0; int pisN = 0; int pisA = 0; int total = 0; int cont = 0;

niño(){
    lock(&m);
    nEsp++;
    while(cont >= 3 && aEsp > 0|| (pisN >= 20 || total >= 40)){
        cond_wait(&cNiño, &m);
    }
    nEsp--;
    pisN++:
    total++;
    cont++;
    unlock(&m);
    baño();
    lock(&m);
    pisN--;
    total--;
    if(nEsp > 0 && cont < 3 && (pisN < 20 && total < 40)){
        cond_signal(&cNiño);
    } else if(aEsp > 0 && cont >= 3 && (pisA < 30 && total < 40)) {
        cond_signal(&cAdulto);
    }
    unlock(&m);
}

adulto(){
    lock(&m);
    aEsp++;
    while(cont < 3 || (pisA >= 30 || total >= 40)){
        cond_wait(&cAdulto, &m);
    }
    aEsp--;
    pisA++;
    total++;
    cont = 0;
    unlock(&m);
    baño();
    lock(&m);
    pisA--;
    total--;
    if(nEsp > 0 && cont < 3 && (pisN < 20 && total < 40)){
        cond_signal(&cNiño);
    } else if(aEsp > 0 && cont >= 3 && (pisA < 30 && total < 40)){
        cond_signal(&cAdulto);
    }
    unlock(&m);
}

mutex_t m;
cond_t cNorte, cSur;
int nN = 0;int nS = 0; int nEsp = 0; int sEsp = 0;

coche_norte(){
    lock(&m);
    nEsp++;
    while(nS > 0 || nN >= 3){
        cond_wait(&cNorte, &m);
    }
    nEsp--
    nN++;
    unlock(&m);
    cruzar_puente();
    lock(&m);
    nN--;
    if(nEsp > 0){
        cond_signal(&cNorte);
    } else if(sEsp > 0 && nN == 0){
        cond_signal(&cSur);
    }
    unlock(&m);
}

coche_sur(){
    lock(&m);
    sEsp++;
    while(nN > 0 || nS >= 3){
        cond_wait(&cSur, &m);
    }
    sEsp--;
    nS++;
    unlock(&m);
    cruzar_puente();
    lock(&m);
    if(nEsp > 0){
        cond_signal(&cNorte);
    } else if(sEsp > 0 && nN == 0){
        cond_signal(&cSur);
    }
    unlock(&m);
}

mutex_t m;
cond_t clleno, cvacio;
int n = 0;

pasajero(){
    lock(&m);
    while(n == 4){
        cond_wait(&cvacio, &m);
    }
    n++
    if(n == 4){
        cond_broadcast(&clleno);
    } else{
        cond_wait(&clleno, &m);
    }
    unlock(&m);
    viaje();
    lock(&m);
    n--;
    if(n==0){
        cond_broadcast(&cvacio, &m);
    }
    unlock(&m);
}


mutex_t m;
cond_t cPA, cPB;
int n = 0; int pesoTotal = 0; int nEspPA = 0; int nEspPB = 0; 

planta_alta(){
    lock(&m);
    nEspPA++;
    while(n >= 6 || pesoTotal + peso >= 500){
        cond_wait(&cPA, &m);
    }
    nEspPA--;
    n++;
    unlock(&m);
    subir_ascensor();
    lock(&m);
    n--;
    pesoTotal = pesoTotal - peso;
    if(nEspPA > 0){
        cond_signal(&cPA);
    } else if(nEspPB > 0){
        cond_signal(&cPB);
    }
    unlock(&m);
}

planta_baja(){
    lock(&m);
    nEspPB++;
    while(n >= 6 || pesoTotal + peso >= 500 || nEspPA > 0){
        cond_wait(&cPB, &m);
    }
    nEspPB--;
    n++;
    unlock(&m);
    subir_ascensor();
    lock(&m);
    n--;
    pesoTotal = pesoTotal - peso;
    if(nEspPA > 0){
        cond_signal(&cPA);
    } else if(nEspPB > 0){
        cond_signal(&cPB);
    }
    unlock(&m);
}

mutex_t m;
cond_t cHP, cHL;
int cont = 0; int nEspHP = 0; int nEspHL = 0;

hilo_pesado(){
    lock(&m);
    nEspHP++;
    while(cont >= 5){
        cond_wait(&cHP, &m);
    }
    nEspHP--;
    cont++;
    unlock(&m);
    usar_recurso();
    lock(&m);
    cont--;
    if(nEspHP > 0 ){
        cond_broadcast(&cHP);
    } else if(nEspHL > 0){
        cond_broadcast(&cHL);
    }
    unlock(&m);
}

hilo_ligero(){
    lock(&m);
    nEspHL++;
    while(cont >= 5 && nEspHP == 0){
        cond_wait(&cHL, &m);
    }
    nEspHL--;
    cont++;
    unlock(&m);
    usar_recurso();
    lock(&m);
    cont--;
    if(nEspHP > 0){
        cond_broadcast(&cHP);
    } else if(nEspHL > 0){
        cond_broadcast(&cHL);
    }
    unlock(&m);
}