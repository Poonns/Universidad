
//EJERCICIO 2 | ZEUS MALLENCO , RAUL MARTINEZ , HECTOR PONS 

#include <iostream>
#include <vector>
#include <string>
#include <unistd.h>
#include <pthread.h>

using namespace std;

char chBase = '-';
char chMarca = '+';
bool terminar = false;

struct dato {
    int col;
    int nCol;
    int inicioUs;
    int intervaloUs;
};

void *hilo(void *arg)
{
    struct dato *d = (struct dato *) arg;
    
    string s;
    s.assign(d->nCol, chBase);
    s[d->col] = chMarca;
    s += '\n';
    
    usleep(d->inicioUs);
    
    while (!terminar)
    {
        cout << s;
        usleep(d->intervaloUs);
    }
    
    return NULL;
}

int main(int argc, char *argv[])
{
    if (argc != 4) {
        cerr << "Uso: " << argv[0] << " <tic_ms> <ntics> <ncol>\n";
        return 1;
    }

    int tic_ms = stoi(argv[1]);
    int ntics = stoi(argv[2]);
    int nCol = stoi(argv[3]);
    
    int ticUs = tic_ms * 1000;
    int tiempo_total_us = ntics * ticUs;

    cout << "tic_ms=" << tic_ms << " ntics=" << ntics << " ncol=" << nCol << "\n";

    struct dato *datos = new struct dato[nCol];
    vector<pthread_t> id(nCol);

    for (int i = 0; i < nCol; ++i) 
    {
        datos[i].col = i;
        datos[i].nCol = nCol;
        datos[i].inicioUs = 25 * 1000 * (i + 1); 
        datos[i].intervaloUs = ticUs;            
        
        pthread_create(&id[i], NULL, hilo, &datos[i]);
    }

    usleep(tiempo_total_us);

    terminar = true;

    for (int i = 0; i < nCol; ++i) 
    {
        pthread_join(id[i], NULL);
    }

    delete[] datos;
    
    cout << "tic_ms=" << tic_ms << " ntics=" << ntics << " ncol=" << nCol << "\n";

    return 0;
}
