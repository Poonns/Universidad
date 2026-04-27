#include <iostream>
#include <string>
#include <vector>
#include <sstream>
#include <unistd.h> //Necesario para execvp
#include <sys/wait.h>

using namespace std;

int main() {
    //Imprimimos el prompt "? " y forzamos su salida a pantalla
    cout<<"? "<<flush;

    //Leemos la línea completa que escribe el usuario
    string linea;
    if (!getline(cin, linea))
        return 0; // ssi hay un error de lectura salimos
    

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
    
    cout<<"padre: creando copia"<< endl;
    pid_t pid = fork();
    
    if (pid < 0) {
        //Si pid es menor que 0 significa que ha flallado al clonar
        perror("Error al hacer fork");
        return 1;
    }
    else if(pid==0){
        cout<< "hijo: a punto de hacer exec"<<endl;
        execvp(args[0], args.data()); //sacrificamos al hijo
        
        //Si llegamos aqui es que el exec ha fallado
        perror("execvp");
        return 1;
    }else{
        cout<<"padre: esperando al hijo"<<endl;
        //Esperamos a que el hijo termine
        wait(NULL);
        cout<<"padre: el hijo ha terminado, finalizando"<< endl;
    }
    
    return 0;
}
