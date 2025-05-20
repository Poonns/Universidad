package Hector.Examenes

import java.io.File;
import java.io.FileNotFoundException;
import java.util.Scanner;

finales.Julio2017;

public class problema2 {
    class Cuentas{
        private class Nodo{
            String nombre;
            double cantidad;
            Nodo sig;

            Nodo(String nombre, double cantidad, Nodo sig){
                this.nombre = nombre;
                this.cantidad = cantidad;
                this.sig = sig;
            }
        }
        private Nodo primero;

        private void insertar(String nombre, double cantidad){
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.nombre.equals(nombre)){
                    actual.cantidad += cantidad;
                }
            }
            Nodo nuevo = new Nodo(nombre, cantidad, primero);
            primero = nuevo;
        }

        private void convertirASaldos(){
            int contador = 0;
            double totalcantidad = 0;
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                totalcantidad += actual.cantidad;
                contador ++;
            }

            double media = totalcantidad / contador;
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                actual.cantidad -= media;
            }
        }
        
        Cuentas(String fichero) throws FileNotFoundException{
            Scanner entrada = new Scanner(new File(fichero));

            while(entrada.hasNextLine()){
                double cantidad = entrada.nextDouble();
                String nombre = entrada.next();
                insertar(nombre, cantidad);
            }
            entrada.close();
            convertirASaldos();
        }

        public void asentarPago(String pagador, String receptor, double cantidad){
            Nodo nodoPagador = null; 
            Nodo antPagador = null;
            Nodo nodoReceptor = null;
            Nodo antReceptor = null;

            Nodo anterior = null;
            
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.nombre.equals(pagador)){
                    nodoPagador = actual;
                    antPagador = anterior;
                }
                if(actual.nombre.equals(receptor)){
                    nodoReceptor = actual;
                    antReceptor = anterior;
                }

                if(nodoPagador != null && nodoReceptor != null){
                    break;
                }

                anterior = actual;

            }

            nodoPagador.cantidad += cantidad;
            nodoReceptor.cantidad -= cantidad;

            if(nodoPagador.cantidad == 0){
                if(antPagador == null){
                    primero = nodoPagador.sig;
                } else{
                    antPagador.sig = nodoPagador.sig;
                }
            }

            if(nodoReceptor.cantidad == 0){
                if(antReceptor == null){
                    primero = nodoReceptor.sig;
                }else{
                    antReceptor.sig = nodoReceptor.sig;
                }
            }
        }
    }
}
