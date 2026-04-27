/*En un parque acuático hay un tobogán gigante muy popular. Para evitar accidentes, el socorrista tiene una regla estricta: no puede haber más de 5 personas lanzándose por el tobogán simultáneamente.
El parque vende dos tipos de entradas: VIP y Normal.
La regla de acceso es la siguiente:
Los clientes VIP tienen prioridad absoluta. Si hay un cliente VIP esperando para tirarse, ningún cliente Normal puede acceder al tobogán, sin importar cuánto tiempo lleve esperando el Normal.
Si no hay clientes VIP esperando, los Normales pueden pasar.
Por supuesto, en ningún caso se debe superar el aforo de 5 personas simultáneas en el tobogán.*/
sem_t s;
sem_t vip;
sem_t normal;

int capacidad = 5;
int nEspVip = 0;
int nEspN = 0;

cliente_vip()
{
    sem_wait(&s);
    if(capacidad > 0)
    {
        capacidad--;
        sem_post(&s);
    } else {
        nEspVip++;
        sem_post(&s);
        sem_wait(&vip);
    }

    tobogan();

    sem_wait(&s);
    capacidad++;

    if(nEspVip > 0){
        nEspVip--;
        capacidad--;
        sem_post(&vip);
    } else if(nEspN > 0){
        nEspN--;
        capacidad--;
        sem_post(&normal);
    }
    sem_post(&s);
}

cliente_normal()
{
    sem_wait(&s);

    if(capacidad > 0 && nEspVip == 0){
        capadidad--;
        sem_post(&s);
    } else {
        nEspN++;
        sem_post(&s);
        sem_wait(&normal);
    }
    tobogan();

    sem_wait(&s);
    capacidad++;

    if(nEspVip > 0){
        nEspVip--;
        capacidad--;
        sem_post(&vip);
    } else if(nEspN > 0){
        nEspN--;
        capacidad--;
        sem_post(&normal);
    }
    sem_post(&s);
}