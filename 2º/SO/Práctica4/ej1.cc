
//EJERCICIO 1 | ZEUS MALLENCO, RAUL MARTINEZ, HECTOR PONS.

#include <iostream>
#include <fstream>
#include <sstream>
#include <cstring>
#include <vector>
#include <list>
#include <pthread.h>

using namespace std;

void *hilo(void *arg)
{
    const char *fichero = (const char *) arg;
    stringstream sout;
    
    ifstream ifs{fichero};
    
    if (not ifs)
    {
        sout << fichero << ": no se puede abrir el fichero para leer\n";
        cout << sout.str();
        
        return NULL; 
    }
    
    list<string> *buf = new list<string>();
    
    string linea;
    while (getline(ifs, linea))
    {
        buf->push_back(linea);
    }
    
    return (void *) buf;
}

int main(int argc, char *argv[])
{
    vector<pthread_t> id;

    if(argc == 1)
    {
        cout << "Hay que pasarle mínimo un argumento \n";
        return 1;   
    }
    
    for (int i = 1; i < argc; ++i) 
    {
        pthread_t t;
        pthread_create(&t, NULL, hilo, argv[i]);
        id.push_back(t);
    }
    
    for (int i = 1; i < argc; ++i) 
    {
        void *r; 
        pthread_join(id[i-1], &r);
        
        if (r == NULL) 
        {
            cout << "fichero \"" << argv[i] << "\": ERROR\n";
        } 
        else 
        {
            cout << "fichero \"" << argv[i] << "\": OK\n";
            
            list<string> *buf = (list<string> *) r;
            
            for (const auto& linea_leida : *buf) 
            {
                cout << linea_leida << "\n";
            }
            
            delete buf; 
        }
    }
    
    return 0;
}
