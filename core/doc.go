package core

/*

core struct

AppCtx---------------------------------------------
		|            |              |             |
        |         TimerMgr     TaskExecutor    Profile
        |
   AppModules--------------------------------------
                     |              |             |
                     |              |   XXX_UserCustomModule
                     |              |
                     |        TransactModule
                     |
                     |
                  NetModule------------------------
                                    |             |
                                    |         Connector-----------------
                                    |                        |         |
                                    |                     Session Socket Connect
                                    |
                                  Acceptor---------------------
                                                  |          |
                                               Session0    Session1..n
*/
