/**
 * Clase para representar alumnos
 * Curso 2025/2026
 */

#include <iostream>
#include <string>
using namespace std;

#ifndef _ALUMNO_H_
#define _ALUMNO_H_

class Alumno
{
    string dni;
    string nombre;
    string apellidos;
    unsigned curso;
    unsigned telefono;
    
    static const unsigned NUM_NOMBRE = 15;
    static string list_nombre[];
    static string list_apellido[];


public:
    /**
     * Constructor por defecto.
     * Crea un alumno vacio.
     */
    Alumno() {};
    /**
     * Crea un alumno para un curso concreto, con el resto de datos aleatorios.
     * @param curs Curso del alumno
     */
    Alumno(unsigned curs);
    /**
     * Devuelve el DNI para usarlo como clave.
     * @return DNI del alumno.
     */
    string getDNI() const 
        {return dni; };
        
    friend ostream & operator<<(ostream &, const Alumno &);
             
};

#endif
