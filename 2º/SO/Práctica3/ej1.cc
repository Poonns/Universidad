#include <iostream>
#include <string>
#include <vector>
#include <sstream>
#include <unistd.h> //Necesario para execvp

using namespace std;

int main() {
    //Imprimimos el prompt "? " y forzamos su salida a pantalla
    cout<<"? "<<flush;

    //Leemos la línea completa que escribe el usuario
    string linea;
    if (!getline(cin, linea))
        return 0; //si hay un error de lectura salimos
    

    //Troceamos la linea para separar el comando de sus argumentos
    istringstream flujo(linea);
    string palabra;
    vector<string> argumentos_str;

    //El flujo extrae palabra por palabra ignorando los espacios automaticamente
    while (flujo >> palabra)
        argumentos_str.push_back(palabra);
    

    //Comprobamos si la línea estaba vacía (solo pulsó INTRO)
    if (argumentos_str.empty())
        return 0; 

    //Preparamos el array tipo C que exige execvp
    vector<char*> args;
    for (size_t i = 0; i < argumentos_str.size(); i++) 
    {
        //Accedemos a cada palabra usando el índice [i]
        args.push_back(const_cast<char*>(argumentos_str[i].c_str())); 
    }
    //El ultimo elemento del array para execvp debe ser siempre nulo
    args.push_back(nullptr); 

    //Llamada al sistema execvp
    //args[0]: nombre del comando (ej: "date")
    //args.data(): entrega el array subyacente compatible con C
    execvp(args[0], args.data());

    //Si el código llega aquí, execvp falla (por ejempplo por un comando inventado)
    perror("Error");
    return 1;
}
